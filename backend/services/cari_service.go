package services

import (
	"github.com/shopspring/decimal"

	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)
var (
	ErrCariNotFound        = errors.New("cari_not_found")
	ErrCariHasTransactions = errors.New("cari_has_transactions")
	ErrDuplicateCariCode   = errors.New("duplicate_cari_code")
	ErrDuplicateTaxNumber  = errors.New("duplicate_tax_number")
	ErrDuplicatePhone      = errors.New("duplicate_phone")
)

type CariService struct{}

func NewCariService() *CariService {
	return &CariService{}
}

type CariInput struct {
	Type        string `json:"type" binding:"required,max=50"`
	Code        string `json:"code" binding:"max=100"`
	Group       string `json:"group" binding:"max=100"`
	Title       string `json:"title" binding:"max=255"`
	Name        string `json:"name" binding:"required,max=255"`
	ContactName string `json:"contact_name" binding:"max=255"`
	TaxOffice   string `json:"tax_office" binding:"max=255"`
	TaxNumber   string `json:"tax_number" binding:"max=50"`
	Email       string `json:"email" binding:"omitempty,email,max=255"`
	Phone       string `json:"phone" binding:"max=50"`
	Landline    string `json:"landline" binding:"max=50"`
	Fax         string `json:"fax" binding:"max=50"`

	// Fatura adresi
	Address    string `json:"address" binding:"max=2000"`
	City       string `json:"city" binding:"max=255"`
	District   string `json:"district" binding:"max=255"`
	PostalCode string `json:"postal_code" binding:"max=20"`
	Country    string `json:"country" binding:"max=255"`

	// Sevk adresi
	ShippingAddress    string `json:"shipping_address" binding:"max=2000"`
	ShippingCity       string `json:"shipping_city" binding:"max=255"`
	ShippingDistrict   string `json:"shipping_district" binding:"max=255"`
	ShippingPostalCode string `json:"shipping_postal_code" binding:"max=20"`
	ShippingCountry    string `json:"shipping_country" binding:"max=255"`

	Currency       string          `json:"currency" binding:"max=10"`
	OpeningBalance decimal.Decimal `json:"opening_balance"` // Deprecated, kept for compatibility
	Note           string          `json:"note" binding:"max=2000"`
}

func (s *CariService) Create(ctx context.Context, in CariInput, createdBy uuid.UUID) (*models.Cari, error) {
	tx := utils.GetDB(ctx, database.DB)
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, err
	}

	// Generate Code if empty
	code := strings.TrimSpace(in.Code)
	if code == "" {
		// Çakışmaya dayanıklı üretim: sayaç ile gerçek veri arasında bir
		// tutarsızlık varsa, boş bir kod bulana kadar sayacı atlatarak ilerle.
		for attempt := 0; attempt < 1000; attempt++ {
			generated, err := utils.GenerateNumberWithSetting(tx, companyID, "cari", "cari_prefix", "ACC")
			if err != nil {
				return nil, err
			}
			var count int64
			if err := tx.Model(&models.Cari{}).Where("company_id = ? AND code = ?", companyID, generated).Count(&count).Error; err != nil {
				return nil, err
			}
			if count == 0 {
				code = generated
				break
			}
		}
		if code == "" {
			return nil, errors.New("benzersiz cari kodu üretilemedi")
		}
	} else {
		// Check uniqueness of code for this company
		var count int64
		if err := tx.Model(&models.Cari{}).Where("company_id = ? AND code = ?", companyID, code).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrDuplicateCariCode
		}
	}

	// Check Uniqueness of TaxNumber (VN/TCKN)
	if strings.TrimSpace(in.TaxNumber) != "" {
		var count int64
		if err := tx.Model(&models.Cari{}).Where("company_id = ? AND tax_number = ?", companyID, in.TaxNumber).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrDuplicateTaxNumber
		}
	}

	// Check Uniqueness of Phone
	if strings.TrimSpace(in.Phone) != "" {
		var count int64
		if err := tx.Model(&models.Cari{}).Where("company_id = ? AND phone = ?", companyID, in.Phone).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrDuplicatePhone
		}
	}

	cariID, _ := uuid.NewV7()
	cari := models.Cari{
		ID:          cariID,
		CompanyID:   companyID,
		Code:        code,
		Type:        in.Type,
		Group:       in.Group,
		Title:       in.Title,
		Name:        in.Name,
		ContactName: in.ContactName,
		TaxOffice:   in.TaxOffice,
		TaxNumber:   in.TaxNumber,
		Email:       strings.ToLower(strings.TrimSpace(in.Email)),
		Phone:       in.Phone,
		Landline:    in.Landline,
		Fax:         in.Fax,

		Address:    in.Address,
		City:       in.City,
		District:   in.District,
		PostalCode: in.PostalCode,
		Country:    in.Country,

		ShippingAddress:    in.ShippingAddress,
		ShippingCity:       in.ShippingCity,
		ShippingDistrict:   in.ShippingDistrict,
		ShippingPostalCode: in.ShippingPostalCode,
		ShippingCountry:    in.ShippingCountry,

		Currency:  in.Currency,
		IsActive:  true,
		Note:      in.Note,
		CreatedBy: &createdBy,
	}

	if cari.Country == "" {
		cari.Country = "Türkiye"
	}
	if cari.Currency == "" {
		cari.Currency = "TRY"
	}

	// Save contact
	if err := tx.Create(&cari).Error; err != nil {
		return nil, err
	}

	if !in.OpeningBalance.IsZero() {
		amount := in.OpeningBalance
		txType := "debit"

		if in.Type == "supplier" {
			amount = in.OpeningBalance.Neg()
			txType = "credit"
		} else {
			if in.OpeningBalance.IsNegative() {
				txType = "credit"
			}
		}

		balanceAfter, err := UpdateCariBalance(tx, cari.ID, cari.Currency, amount)
		if err != nil {
			return nil, err
		}

		txID, _ := uuid.NewV7()
		cTx := models.CariTransaction{
			ID:           txID,
			CompanyID:    companyID,
			CariID:       cari.ID,
			Date:         utils.NowIn(ctx),
			Type:         txType,
			SourceType:   "opening_balance",
			Currency:     cari.Currency,
			Description:  "Devir Bakiyesi (Açılış)",
			Amount:       in.OpeningBalance.Abs(),
			BalanceAfter: balanceAfter,
			CreatedBy:    &createdBy,
		}
		if err := tx.Create(&cTx).Error; err != nil {
			return nil, err
		}
	}

	if err := WriteAuditLog(ctx, tx, "cari", cari.ID, "create", createdBy, cari.Name); err != nil {
		return nil, err
	}

	// Preload Balances before returning
	if err := tx.Preload("Balances").First(&cari, cari.ID).Error; err != nil {
		return nil, err
	}

	return &cari, nil
}

func (s *CariService) Update(ctx context.Context, id uuid.UUID, in CariInput, userID uuid.UUID) (*models.Cari, error) {
	tx := utils.GetDB(ctx, database.DB)

	var cari models.Cari
	if err := tx.First(&cari, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCariNotFound
		}
		return nil, err
	}

	// Check Uniqueness of TaxNumber (VN/TCKN)
	if strings.TrimSpace(in.TaxNumber) != "" && in.TaxNumber != cari.TaxNumber {
		var count int64
		if err := tx.Model(&models.Cari{}).Where("company_id = ? AND tax_number = ? AND id != ?", cari.CompanyID, in.TaxNumber, id).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrDuplicateTaxNumber
		}
	}

	// Check Uniqueness of Phone
	if strings.TrimSpace(in.Phone) != "" && in.Phone != cari.Phone {
		var count int64
		if err := tx.Model(&models.Cari{}).Where("company_id = ? AND phone = ? AND id != ?", cari.CompanyID, in.Phone, id).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrDuplicatePhone
		}
	}

	// Code and OpeningBalance cannot be updated directly
	cari.Type = in.Type
	cari.Group = in.Group
	cari.Title = in.Title
	cari.Name = in.Name
	cari.ContactName = in.ContactName
	cari.TaxOffice = in.TaxOffice
	cari.TaxNumber = in.TaxNumber
	cari.Email = strings.ToLower(strings.TrimSpace(in.Email))
	cari.Phone = in.Phone
	cari.Landline = in.Landline
	cari.Fax = in.Fax

	cari.Address = in.Address
	cari.City = in.City
	cari.District = in.District
	cari.PostalCode = in.PostalCode
	cari.Country = in.Country

	cari.ShippingAddress = in.ShippingAddress
	cari.ShippingCity = in.ShippingCity
	cari.ShippingDistrict = in.ShippingDistrict
	cari.ShippingPostalCode = in.ShippingPostalCode
	cari.ShippingCountry = in.ShippingCountry

	cari.Currency = in.Currency
	cari.Note = in.Note

	if err := tx.Save(&cari).Error; err != nil {
		return nil, err
	}

	if err := WriteAuditLog(ctx, tx, "cari", cari.ID, "update", userID, cari.Name); err != nil {
		return nil, err
	}

	return &cari, nil
}

func (s *CariService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	tx := utils.GetDB(ctx, database.DB)

	var cari models.Cari
	if err := tx.First(&cari, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCariNotFound
		}
		return err
	}

	// Check if contact has ledger transactions
	var count int64
	if err := tx.Model(&models.CariTransaction{}).Where("cari_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrCariHasTransactions
	}

	if err := WriteAuditLog(ctx, tx, "cari", cari.ID, "delete", userID, cari.Name); err != nil {
		return err
	}

	return tx.Delete(&cari).Error
}

func (s *CariService) GetByID(ctx context.Context, id uuid.UUID) (*models.Cari, error) {
	tx := utils.GetDB(ctx, database.DB)

	var cari models.Cari
	if err := tx.Preload("Balances").Preload("Persons").First(&cari, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCariNotFound
		}
		return nil, err
	}
	return &cari, nil
}

func (s *CariService) List(ctx context.Context, page, limit int, query, sort string, filters map[string]string) ([]models.Cari, int64, error) {
	tx := utils.GetDB(ctx, database.DB)

	var caris []models.Cari
	var total int64

	dbQuery := tx.Model(&models.Cari{})

	if query != "" {
		q := "%" + query + "%"
		dbQuery = dbQuery.Where("name ILIKE ? OR title ILIKE ? OR code ILIKE ?", q, q, q)
	}

	// Apply type filter
	if tFilter, exists := filters["type"]; exists && tFilter != "" {
		dbQuery = dbQuery.Where("type = ?", tFilter)
	}

	// Apply group filter
	if gFilter, exists := filters["group"]; exists && gFilter != "" {
		dbQuery = dbQuery.Where("\"group\" = ?", gFilter)
	}

	// Apply active status filter
	if activeFilter, exists := filters["is_active"]; exists && activeFilter != "" {
		dbQuery = dbQuery.Where("is_active = ?", activeFilter == "true")
	}

	// Count total records
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply Sorting
	if sort != "" {
		dbQuery = dbQuery.Order(sort)
	} else {
		dbQuery = dbQuery.Order("created_at DESC")
	}

	// Apply Pagination and Preload Balances
	offset := (page - 1) * limit
	if err := dbQuery.Preload("Balances").Offset(offset).Limit(limit).Find(&caris).Error; err != nil {
		return nil, 0, err
	}

	return caris, total, nil
}

func (s *CariService) GetTransactions(ctx context.Context, cariID uuid.UUID, page, limit int) ([]models.CariTransaction, int64, error) {
	tx := utils.GetDB(ctx, database.DB)

	var txs []models.CariTransaction
	var total int64

	dbQuery := tx.Model(&models.CariTransaction{}).Where("cari_id = ?", cariID)

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sort ledger ascending to show chronologically correct rolling balances
	offset := (page - 1) * limit
	if err := dbQuery.Order("date ASC, created_at ASC").Offset(offset).Limit(limit).Find(&txs).Error; err != nil {
		return nil, 0, err
	}

	return txs, total, nil
}

func (s *CariService) GetSummary(ctx context.Context) (map[string]interface{}, error) {
	tx := utils.GetDB(ctx, database.DB)
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}

	type CurrencySummary struct {
		Currency    string  `gorm:"column:currency"`
		Receivables float64 `gorm:"column:receivables"`
		Payables    float64 `gorm:"column:payables"`
	}

	var summaries []CurrencySummary

	query := `
		SELECT 
			cb.currency,
			COALESCE(SUM(CASE WHEN cb.balance > 0 THEN cb.balance ELSE 0 END), 0) as receivables,
			COALESCE(SUM(CASE WHEN cb.balance < 0 THEN ABS(cb.balance) ELSE 0 END), 0) as payables
		FROM caris c
		JOIN cari_balances cb ON cb.cari_id = c.id
		WHERE c.company_id = ? AND c.is_active = true
		GROUP BY cb.currency
	`
	if err := tx.Raw(query, companyIDStr).Scan(&summaries).Error; err != nil {
		return nil, err
	}

	type CurrencyAmount struct {
		Currency string          `json:"currency"`
		Amount   decimal.Decimal `json:"amount"`
	}

	var totalReceivables []CurrencyAmount
	var totalPayables []CurrencyAmount
	var netBalances []CurrencyAmount

	for _, s := range summaries {
		receivablesDec := decimal.NewFromFloat(s.Receivables)
		payablesDec := decimal.NewFromFloat(s.Payables)
		netDec := receivablesDec.Sub(payablesDec)

		totalReceivables = append(totalReceivables, CurrencyAmount{Currency: s.Currency, Amount: receivablesDec})
		totalPayables = append(totalPayables, CurrencyAmount{Currency: s.Currency, Amount: payablesDec})
		netBalances = append(netBalances, CurrencyAmount{Currency: s.Currency, Amount: netDec})
	}

	return map[string]interface{}{
		"total_receivables": totalReceivables,
		"total_payables":    totalPayables,
		"net_balances":      netBalances,
	}, nil
}

func (s *CariService) GetFinancialSummary(ctx context.Context, cariID uuid.UUID) (map[string]interface{}, error) {
	tx := utils.GetDB(ctx, database.DB)

	type CurrencyTotal struct {
		Currency string  `gorm:"column:currency" json:"currency"`
		Total    float64 `gorm:"column:total" json:"total"`
	}

	var sales []CurrencyTotal
	var purchases []CurrencyTotal
	var collections []CurrencyTotal
	var payments []CurrencyTotal

	// Total Sales
	if err := tx.Model(&models.Invoice{}).
		Select("currency, COALESCE(SUM(total), 0) as total").
		Where("cari_id = ? AND type = 'sales' AND status NOT IN ('draft', 'canceled')", cariID).
		Group("currency").
		Scan(&sales).Error; err != nil {
		return nil, err
	}

	// Total Purchases
	if err := tx.Model(&models.Invoice{}).
		Select("currency, COALESCE(SUM(total), 0) as total").
		Where("cari_id = ? AND type = 'purchase' AND status NOT IN ('draft', 'canceled')", cariID).
		Group("currency").
		Scan(&purchases).Error; err != nil {
		return nil, err
	}

	// Total Collections (Tahsilat)
	if err := tx.Model(&models.Payment{}).
		Select("currency, COALESCE(SUM(amount), 0) as total").
		Where("cari_id = ? AND type = 'collection' AND status != 'canceled'", cariID).
		Group("currency").
		Scan(&collections).Error; err != nil {
		return nil, err
	}

	// Total Payments (Ödeme)
	if err := tx.Model(&models.Payment{}).
		Select("currency, COALESCE(SUM(amount), 0) as total").
		Where("cari_id = ? AND type = 'payment' AND status != 'canceled'", cariID).
		Group("currency").
		Scan(&payments).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"sales":       sales,
		"purchases":   purchases,
		"collections": collections,
		"payments":    payments,
	}, nil
}

// UpdateCariBalance is a helper function to safely update the balance of a specific currency for a Cari
func UpdateCariBalance(tx *gorm.DB, cariID uuid.UUID, currency string, amount decimal.Decimal) (decimal.Decimal, error) {
	var cb models.CariBalance
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("cari_id = ? AND currency = ?", cariID, currency).First(&cb).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cb = models.CariBalance{
				ID:       uuid.New(),
				CariID:   cariID,
				Currency: currency,
				Balance:  decimal.Zero,
			}
			if err := tx.Create(&cb).Error; err != nil {
				return decimal.Zero, err
			}
		} else {
			return decimal.Zero, err
		}
	}

	cb.Balance = cb.Balance.Add(amount)
	if err := tx.Save(&cb).Error; err != nil {
		return decimal.Zero, err
	}

	return cb.Balance, nil
}

func (s *CariService) GetNextCode(ctx context.Context) (map[string]interface{}, error) {
	tx := utils.GetDB(ctx, database.DB)
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, err
	}

	code, err := utils.PreviewNumberWithSetting(tx, companyID, "cari", "cari_prefix", "ACC")
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"next_code": code}, nil
}

func (s *CariService) AddPerson(ctx context.Context, cariID uuid.UUID, name, title, phone, email string) (*models.CariPerson, error) {
	tx := utils.GetDB(ctx, database.DB)
	person := &models.CariPerson{
		ID:     uuid.New(),
		CariID: cariID,
		Name:   name,
		Title:  title,
		Phone:  phone,
		Email:  strings.ToLower(strings.TrimSpace(email)),
	}

	if err := tx.Create(person).Error; err != nil {
		return nil, err
	}
	return person, nil
}

func (s *CariService) UpdatePerson(ctx context.Context, personID uuid.UUID, name, title, phone, email string) (*models.CariPerson, error) {
	tx := utils.GetDB(ctx, database.DB)

	var person models.CariPerson
	if err := tx.First(&person, personID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("person_not_found")
		}
		return nil, err
	}

	person.Name = name
	person.Title = title
	person.Phone = phone
	person.Email = strings.ToLower(strings.TrimSpace(email))

	if err := tx.Save(&person).Error; err != nil {
		return nil, err
	}
	return &person, nil
}

func (s *CariService) RemovePerson(ctx context.Context, personID uuid.UUID) error {
	tx := utils.GetDB(ctx, database.DB)
	return tx.Delete(&models.CariPerson{}, "id = ?", personID).Error
}
