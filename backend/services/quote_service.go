package services

import (
	"github.com/shopspring/decimal"

	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)

var (
	ErrQuoteNotFound         = errors.New("quote_not_found")
	ErrQuoteNotEditable      = errors.New("quote_not_editable")  // Once accepted/converted, cannot edit
	ErrQuoteNotDeletable     = errors.New("quote_not_deletable") // Once accepted/converted, cannot delete
	ErrQuoteAlreadyConverted = errors.New("quote_already_converted")
	ErrCariNotFoundForQuote  = errors.New("cari_not_found")
)

var allowedQuoteTransitions = map[string][]string{
	"draft":     {"sent"},
	"sent":      {"accepted", "rejected", "expired"},
	"accepted":  {},
	"rejected":  {"sent"},
	"expired":   {"sent"},
	"converted": {},
}

func canQuoteTransition(current, next string) bool {
	if current == next {
		return true
	}
	allowed, ok := allowedQuoteTransitions[current]
	if !ok {
		return false
	}
	for _, status := range allowed {
		if status == next {
			return true
		}
	}
	return false
}

type QuoteService struct{}

func NewQuoteService() *QuoteService {
	return &QuoteService{}
}

type QuoteItemInput struct {
	ProductID    *uuid.UUID      `json:"product_id"`
	Description  string          `json:"description" binding:"max=2000"`
	Quantity     decimal.Decimal `json:"quantity" binding:"required"`
	Unit         string          `json:"unit" binding:"max=50"`
	UnitPrice    decimal.Decimal `json:"unit_price" binding:"required"`
	DiscountRate decimal.Decimal `json:"discount_rate"`
	TaxRate      decimal.Decimal `json:"tax_rate"`
	Currency       string          `json:"currency" binding:"max=10"`
	ExchangeRate   decimal.Decimal `json:"exchange_rate"`
	ExchangeRateOp string          `json:"exchange_rate_op" binding:"max=5"`
}

type QuoteInput struct {
	CariID     uuid.UUID        `json:"cari_id" binding:"required"`
	Number     string           `json:"number" binding:"max=100"`
	Date       time.Time        `json:"date" binding:"required"`
	ExpiryDate time.Time        `json:"expiry_date" binding:"required"`
	Currency   string           `json:"currency" binding:"max=10"`
	Note       string           `json:"note" binding:"max=2000"`
	Status     string           `json:"status" binding:"required,oneof=draft sent accepted rejected expired converted"`
	Items      []QuoteItemInput `json:"items" binding:"required,min=1"`
}

func (s *QuoteService) Create(ctx context.Context, in QuoteInput, createdBy uuid.UUID) (*models.Quote, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, err
	}

	var quote models.Quote

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		// 1. Verify Cari exists
		var cari models.Cari
		if err := txTenant.First(&cari, "id = ?", in.CariID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCariNotFoundForQuote
			}
			return err
		}

		// 1b. Fetch company's default currency (multi-currency line conversion anchor)
		var company models.Company
		if err := database.SystemDB.First(&company, "id = ?", companyID).Error; err != nil {
			return err
		}
		defaultCurrency := company.Currency
		if defaultCurrency == "" {
			defaultCurrency = "TRY"
		}

		// 2. Generate Number if empty
		number := strings.TrimSpace(in.Number)
		if number == "" {
			// Çakışmaya dayanıklı üretim: sayaç ile gerçek veri arasında bir
			// tutarsızlık varsa, boş bir numara bulana kadar sayacı atlatarak ilerle.
			for attempt := 0; attempt < 1000; attempt++ {
				generated, err := utils.GenerateNumberWithSetting(txTenant, companyID, "quote", "quote_prefix", "PRO")
				if err != nil {
					return err
				}
				var count int64
				if err := txTenant.Model(&models.Quote{}).Where("company_id = ? AND number = ?", companyID, generated).Count(&count).Error; err != nil {
					return err
				}
				if count == 0 {
					number = generated
					break
				}
			}
			if number == "" {
				return errors.New("benzersiz teklif numarası üretilemedi")
			}
		} else {
			// Verify number uniqueness
			var count int64
			if err := txTenant.Model(&models.Quote{}).Where("company_id = ? AND number = ?", companyID, number).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return errors.New("duplicate_quote_number")
			}
		}

		// 3. Calculate Totals
		var subtotal, discountTotal, taxTotal, total decimal.Decimal
		quoteItems := make([]models.QuoteItem, len(in.Items))

		for i, item := range in.Items {
			qty := item.Quantity
			price := item.UnitPrice
			discRate := item.DiscountRate
			tRate := item.TaxRate

			itemCurrency := item.Currency
			if itemCurrency == "" {
				itemCurrency = defaultCurrency
			}
			itemRate := item.ExchangeRate
			itemOp := item.ExchangeRateOp
			if itemRate.IsZero() {
				// Kullanıcı kur girmediyse admin Para Birimi tablosundaki güncel
				// kuru/işlemi kullan (manuel override edilmediğinde tek kur kaynağı).
				itemRate, itemOp = GetCurrencyRateToBase(txTenant, defaultCurrency, itemCurrency)
			}
			if itemOp == "" {
				itemOp = "*"
			}

			lineSub := qty.Mul(price)
			lineDisc := lineSub.Mul(discRate).Div(decimal.NewFromInt(100))
			lineTaxable := lineSub.Sub(lineDisc)
			lineTax := lineTaxable.Mul(tRate).Div(decimal.NewFromInt(100))
			lineTot := lineTaxable.Add(lineTax)

			// Quote-seviyesi toplamlar her zaman şirketin varsayılan dövizinde:
			// her satır kendi kuru/işaretiyle çevrilip toplanır.
			subtotal = subtotal.Add(convertToDefaultCurrency(lineSub, itemCurrency, defaultCurrency, itemRate, itemOp))
			discountTotal = discountTotal.Add(convertToDefaultCurrency(lineDisc, itemCurrency, defaultCurrency, itemRate, itemOp))
			taxTotal = taxTotal.Add(convertToDefaultCurrency(lineTax, itemCurrency, defaultCurrency, itemRate, itemOp))
			total = total.Add(convertToDefaultCurrency(lineTot, itemCurrency, defaultCurrency, itemRate, itemOp))

			itemID, _ := uuid.NewV7()
			quoteItems[i] = models.QuoteItem{
				ID:           itemID,
				CompanyID:    companyID,
				ProductID:    item.ProductID,
				Description:  item.Description,
				Quantity:     item.Quantity,
				Unit:         item.Unit,
				UnitPrice:    item.UnitPrice,
				DiscountRate: item.DiscountRate,
				TaxRate:      item.TaxRate,
				LineTotal:    lineTot, // satırın kendi dövizinde
				Currency:       itemCurrency,
				ExchangeRate:   itemRate,
				ExchangeRateOp: itemOp,
			}
		}

		quoteID, _ := uuid.NewV7()
		publicToken := uuid.NewString()

		// İstenen statü 'sent' olsa da teklif önce 'draft' olarak oluşturulur;
		// mail başarıyla gittiğinde sendQuoteEmailTx statüyü 'sent'e taşır.
		requestedStatus := in.Status

		quote = models.Quote{
			ID:            quoteID,
			CompanyID:     companyID,
			CariID:        in.CariID,
			Number:        number,
			Date:          in.Date,
			ExpiryDate:    in.ExpiryDate,
			// Belge başlığı ve toplamları her zaman base (şirket varsayılan)
			// dövizindedir; satırlar kendi dövizinde olabilir ve base'e çevrilir.
			// Faturayla tutarlı: başlık dövizi ile toplam dövizi artık çelişmez.
			Currency:      defaultCurrency,
			ExchangeRate:  decimal.NewFromFloat(1.0),
			Subtotal:      subtotal,
			DiscountTotal: discountTotal,
			TaxTotal:      taxTotal,
			Total:         total,
			Status:        "draft",
			PublicToken:   publicToken,
			Note:          in.Note,
			Items:         quoteItems,
			CreatedBy:     &createdBy,
		}

		// Save Quote
		if err := txTenant.Create(&quote).Error; err != nil {
			return err
		}

		if requestedStatus == "sent" {
			if err := sendQuoteEmailTx(ctx, txTenant, &quote, &cari, &company, createdBy); err != nil {
				return err
			}
		} else if requestedStatus != "draft" {
			return fmt.Errorf("statü '%s' ile teklif oluşturulamaz", requestedStatus)
		}

		if err := WriteAuditLog(ctx, txTenant, "quote", quote.ID, "create", createdBy, quote.Number); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &quote, nil
}

func (s *QuoteService) Update(ctx context.Context, id uuid.UUID, in QuoteInput, updatedBy uuid.UUID) (*models.Quote, error) {
	var quote models.Quote

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		// 1. Fetch current quote with Items
		if err := txTenant.Preload("Items").First(&quote, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrQuoteNotFound
			}
			return err
		}

		// 2. Only allow editing non-final quotes
		if quote.Status == "converted" || quote.Status == "accepted" {
			return ErrQuoteNotEditable
		}

		// 3. Verify Cari exists
		var cari models.Cari
		if err := txTenant.First(&cari, "id = ?", in.CariID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCariNotFoundForQuote
			}
			return err
		}

		// 3b. Fetch company's default currency (multi-currency line conversion anchor)
		var company models.Company
		if err := database.SystemDB.First(&company, "id = ?", quote.CompanyID).Error; err != nil {
			return err
		}
		defaultCurrency := company.Currency
		if defaultCurrency == "" {
			defaultCurrency = "TRY"
		}

		// 4. Recalculate totals and clear old items
		if err := txTenant.Where("quote_id = ?", id).Delete(&models.QuoteItem{}).Error; err != nil {
			return err
		}

		var subtotal, discountTotal, taxTotal, total decimal.Decimal
		quoteItems := make([]models.QuoteItem, len(in.Items))

		for i, item := range in.Items {
			qty := item.Quantity
			price := item.UnitPrice
			discRate := item.DiscountRate
			tRate := item.TaxRate

			itemCurrency := item.Currency
			if itemCurrency == "" {
				itemCurrency = defaultCurrency
			}
			itemRate := item.ExchangeRate
			itemOp := item.ExchangeRateOp
			if itemRate.IsZero() {
				// Kullanıcı kur girmediyse admin Para Birimi tablosundaki güncel
				// kuru/işlemi kullan (manuel override edilmediğinde tek kur kaynağı).
				itemRate, itemOp = GetCurrencyRateToBase(txTenant, defaultCurrency, itemCurrency)
			}
			if itemOp == "" {
				itemOp = "*"
			}

			lineSub := qty.Mul(price)
			lineDisc := lineSub.Mul(discRate).Div(decimal.NewFromInt(100))
			lineTaxable := lineSub.Sub(lineDisc)
			lineTax := lineTaxable.Mul(tRate).Div(decimal.NewFromInt(100))
			lineTot := lineTaxable.Add(lineTax)

			subtotal = subtotal.Add(convertToDefaultCurrency(lineSub, itemCurrency, defaultCurrency, itemRate, itemOp))
			discountTotal = discountTotal.Add(convertToDefaultCurrency(lineDisc, itemCurrency, defaultCurrency, itemRate, itemOp))
			taxTotal = taxTotal.Add(convertToDefaultCurrency(lineTax, itemCurrency, defaultCurrency, itemRate, itemOp))
			total = total.Add(convertToDefaultCurrency(lineTot, itemCurrency, defaultCurrency, itemRate, itemOp))

			itemID, _ := uuid.NewV7()
			quoteItems[i] = models.QuoteItem{
				ID:           itemID,
				CompanyID:    quote.CompanyID,
				QuoteID:      id,
				ProductID:    item.ProductID,
				Description:  item.Description,
				Quantity:     item.Quantity,
				Unit:         item.Unit,
				UnitPrice:    item.UnitPrice,
				DiscountRate: item.DiscountRate,
				TaxRate:      item.TaxRate,
				LineTotal:    lineTot,
				Currency:       itemCurrency,
				ExchangeRate:   itemRate,
				ExchangeRateOp: itemOp,
			}
		}

		requestedStatus := in.Status

		quote.CariID = in.CariID
		quote.Date = in.Date
		quote.ExpiryDate = in.ExpiryDate
		quote.Note = in.Note
		quote.Status = "draft"
		quote.Subtotal = subtotal
		quote.DiscountTotal = discountTotal
		quote.TaxTotal = taxTotal
		quote.Total = total
		quote.Items = quoteItems
		quote.UpdatedBy = &updatedBy
		// Başlık ve toplamlar base dövizinde (faturayla tutarlı).
		quote.Currency = defaultCurrency
		quote.ExchangeRate = decimal.NewFromInt(1)

		// Save Quote and new Items
		if err := txTenant.Save(&quote).Error; err != nil {
			return err
		}

		if requestedStatus == "sent" {
			if err := sendQuoteEmailTx(ctx, txTenant, &quote, &cari, &company, updatedBy); err != nil {
				return err
			}
		} else if requestedStatus != "draft" {
			return fmt.Errorf("statü '%s' ile teklif güncellenemez", requestedStatus)
		}

		if err := WriteAuditLog(ctx, txTenant, "quote", quote.ID, "update", updatedBy, quote.Number); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &quote, nil
}

func (s *QuoteService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		var quote models.Quote
		if err := txTenant.First(&quote, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrQuoteNotFound
			}
			return err
		}

		// Only allow deleting drafts/non-finalized quotes
		if quote.Status == "converted" || quote.Status == "accepted" {
			return ErrQuoteNotDeletable
		}

		if err := WriteAuditLog(ctx, txTenant, "quote", quote.ID, "delete", userID, quote.Number); err != nil {
			return err
		}

		return txTenant.Delete(&quote).Error
	})
}

func (s *QuoteService) GetByID(ctx context.Context, id uuid.UUID) (*models.Quote, error) {
	tx := utils.GetDB(ctx, database.DB)

	// Automatically transition if expired
	tx.Model(&models.Quote{}).Where("id = ? AND status = 'sent' AND expiry_date < ?", id, time.Now()).Update("status", "expired")

	var quote models.Quote
	if err := tx.Preload("Items").Preload("CreatedByUser").Preload("UpdatedByUser").First(&quote, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrQuoteNotFound
		}
		return nil, err
	}
	return &quote, nil
}

func (s *QuoteService) List(ctx context.Context, page, limit int, query, sort string, filters map[string]string) ([]models.Quote, int64, error) {
	tx := utils.GetDB(ctx, database.DB)

	// Automatically transition expired quotes before listing
	tx.Model(&models.Quote{}).Where("status = 'sent' AND expiry_date < ?", time.Now()).Update("status", "expired")

	var quotes []models.Quote
	var total int64

	dbQuery := tx.Model(&models.Quote{})

	if query != "" {
		q := "%" + strings.ToLower(query) + "%"
		dbQuery = dbQuery.Where("LOWER(number) LIKE ? OR LOWER(note) LIKE ?", q, q)
	}

	if statusFilter, exists := filters["status"]; exists && statusFilter != "" {
		dbQuery = dbQuery.Where("status = ?", statusFilter)
	}

	if cariFilter, exists := filters["cari_id"]; exists && cariFilter != "" {
		dbQuery = dbQuery.Where("cari_id = ?", cariFilter)
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if sort != "" {
		dbQuery = dbQuery.Order(sort)
	} else {
		dbQuery = dbQuery.Order("date DESC, created_at DESC")
	}

	offset := (page - 1) * limit
	if err := dbQuery.Preload("CreatedByUser").Preload("UpdatedByUser").Offset(offset).Limit(limit).Find(&quotes).Error; err != nil {
		return nil, 0, err
	}

	return quotes, total, nil
}

func (s *QuoteService) UpdateStatus(ctx context.Context, id uuid.UUID, status string, userID uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		var quote models.Quote
		if err := txTenant.First(&quote, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrQuoteNotFound
			}
			return err
		}

		if quote.Status == "converted" {
			return errors.New("cannot_change_status_of_converted_quote")
		}

		if status == "converted" {
			return errors.New("must_use_convert_endpoint_to_convert_quote")
		}

		if !canQuoteTransition(quote.Status, status) {
			return errors.New("invalid_status_transition")
		}

		quote.Status = status
		quote.UpdatedBy = &userID

		if err := txTenant.Save(&quote).Error; err != nil {
			return err
		}

		if err := WriteAuditLog(ctx, txTenant, "quote", quote.ID, "update_status", userID, quote.Number+" - "+status); err != nil {
			return err
		}

		return nil
	})
}

// Convert converts a quote to a draft Sales Invoice
func (s *QuoteService) Convert(ctx context.Context, id uuid.UUID, createdBy uuid.UUID) (*models.Invoice, error) {
	var invoice models.Invoice

	runInTx := func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		// 1. Fetch Quote with Items and lock row
		var quote models.Quote
		if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(&quote, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrQuoteNotFound
			}
			return err
		}

		companyID := quote.CompanyID

		// 2. Check if already converted
		if quote.Status == "converted" || quote.ConvertedInvoiceID != nil {
			return ErrQuoteAlreadyConverted
		}

		// 3. Generate sequential draft sales invoice number
		invoiceNumber, err := utils.GenerateNumberWithSetting(txTenant, companyID, "invoice_sales", "invoice_prefix", "INV-S")
		if err != nil {
			return err
		}

		// 4. Map Quote Items to Invoice Items
		invoiceItems := make([]models.InvoiceItem, len(quote.Items))
		for i, qItem := range quote.Items {
			itemID, _ := uuid.NewV7()
			invoiceItems[i] = models.InvoiceItem{
				ID:             itemID,
				CompanyID:      companyID,
				ProductID:      qItem.ProductID,
				Description:    qItem.Description,
				Quantity:       qItem.Quantity,
				Unit:           qItem.Unit,
				UnitPrice:      qItem.UnitPrice,
				DiscountRate:   qItem.DiscountRate,
				TaxRate:        qItem.TaxRate,
				LineTotal:      qItem.LineTotal,
				Currency:       qItem.Currency,
				ExchangeRate:   qItem.ExchangeRate,
				ExchangeRateOp: qItem.ExchangeRateOp,
			}
		}

		now := utils.NowIn(ctx)

		// 5. Create Draft (or Sent, if converted automatically via Public) Sales Invoice
		invoiceID, _ := uuid.NewV7()
		invoice = models.Invoice{
			ID:            invoiceID,
			CompanyID:     companyID,
			CariID:        quote.CariID,
			Type:          "sales",
			Number:        invoiceNumber,
			Date:          now,
			DueDate:       now.AddDate(0, 0, 7), // Default 7 days due date
			Currency:      quote.Currency,
			ExchangeRate:  quote.ExchangeRate,
			Subtotal:      quote.Subtotal,
			DiscountTotal: quote.DiscountTotal,
			TaxTotal:      quote.TaxTotal,
			Total:         quote.Total,
			PaidTotal:     decimal.Zero,
			Status:        "sent", // Directly marked as sent since customer accepted it
			SentAt:        &now,
			LastSentAt:    &now,
			PublicToken:   uuid.NewString(),
			Note:          "Tekliften dönüştürüldü. Teklif No: " + quote.Number + "\n" + quote.Note,
			Items:         invoiceItems,
			CreatedBy:     &createdBy,
		}

		if err := txTenant.Create(&invoice).Error; err != nil {
			return err
		}

		// 6. Update Quote status and link invoice
		quote.Status = "converted"
		quote.ConvertedInvoiceID = &invoice.ID
		quote.UpdatedBy = &createdBy

		if err := txTenant.Save(&quote).Error; err != nil {
			return err
		}

		if err := WriteAuditLog(ctx, txTenant, "quote", quote.ID, "convert", createdBy, quote.Number); err != nil {
			return err
		}
		if err := WriteAuditLog(ctx, txTenant, "invoice", invoice.ID, "create", createdBy, invoice.Number); err != nil {
			return err
		}

		return nil
	}

	var err error
	if existingTx, ok := ctx.Value(utils.TxKey).(*gorm.DB); ok && existingTx != nil {
		err = runInTx(existingTx)
	} else {
		err = database.DB.Transaction(runInTx)
	}

	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (s *QuoteService) SendQuote(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		var quote models.Quote
		if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&quote, "id = ?", id).Error; err != nil {
			return err
		}

		if !canQuoteTransition(quote.Status, "sent") {
			return errors.New("fatura bu durumda gönderilemez (sadece taslak veya reddedilmiş vb.)")
		}

		var cari models.Cari
		if err := txTenant.First(&cari, "id = ?", quote.CariID).Error; err != nil {
			return err
		}

		var company models.Company
		if err := database.SystemDB.First(&company, "id = ?", quote.CompanyID).Error; err != nil {
			return err
		}

		return sendQuoteEmailTx(ctx, txTenant, &quote, &cari, &company, userID)
	})
}

// sendQuoteEmailTx, teklifin mail gönderim+statü güncelleme mantığını taşır.
// Hem SendQuote (manuel "Gönder" butonu) hem Create/Update (durum doğrudan
// 'sent' seçildiğinde) tarafından, ikisi de kendi transaction'larının içinde
// olacak şekilde çağrılır; mail başarısız olursa hata döner ve çağıran
// transaction'ı rollback eder (status hiç 'sent' yazılmaz).
func sendQuoteEmailTx(ctx context.Context, txTenant *gorm.DB, quote *models.Quote, cari *models.Cari, company *models.Company, userID uuid.UUID) error {
	if cari.Email == "" {
		return errors.New("cari_email_not_found")
	}

	// Eski kayıtlarda (PublicToken alanı eklenmeden önce oluşturulmuş
	// tekliflerde) bu alan boş olabilir; boş token'la link üretip kırık
	// bir mail göndermek yerine burada token'ı tamamlayıp devam ediyoruz.
	if quote.PublicToken == "" {
		quote.PublicToken = uuid.NewString()
		if err := txTenant.Model(quote).Update("public_token", quote.PublicToken).Error; err != nil {
			return err
		}
	}

	// App URL for the public link
	appURL := os.Getenv("PUBLIC_APP_URL")
	if appURL == "" {
		appURL = "http://localhost:5173" // fallback
	}
	publicLink := appURL + "/quote/" + quote.PublicToken

	companySummary := utils.CompanySummary{
		Name:    company.Name,
		Phone:   company.Phone,
		Email:   company.Email,
		Address: company.Address,
	}
	totalStr := quote.Total.StringFixed(2) + " " + quote.Currency

	lines := make([]utils.DocumentLine, 0, len(quote.Items))
	for _, item := range quote.Items {
		// Satır tutarları satırın KENDİ dövizinde; satır dövizi ile etiketlenir.
		lineCur := item.Currency
		if lineCur == "" {
			lineCur = quote.Currency
		}
		lines = append(lines, utils.DocumentLine{
			Description: item.Description,
			Quantity:    item.Quantity.StringFixed(2),
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice.StringFixed(2) + " " + lineCur,
			Total:       item.LineTotal.StringFixed(2) + " " + lineCur,
		})
	}
	custName := cari.Title
	if custName == "" {
		custName = cari.Name
	}
	discountText := ""
	if quote.DiscountTotal.IsPositive() {
		discountText = "-" + quote.DiscountTotal.StringFixed(2) + " " + quote.Currency
	}
	doc := utils.DocumentSummary{
		Number:       quote.Number,
		Date:         quote.Date.Format("02.01.2006"),
		DueDate:      quote.ExpiryDate.Format("02.01.2006"),
		Customer:     utils.DocumentParty{Name: custName, Address: cari.Address},
		Lines:        lines,
		Subtotal:     quote.Subtotal.StringFixed(2) + " " + quote.Currency,
		DiscountText: discountText,
		TaxText:      quote.TaxTotal.StringFixed(2) + " " + quote.Currency,
		Total:        quote.Total.StringFixed(2) + " " + quote.Currency,
		Currency:     quote.Currency,
	}

	emailBody := utils.QuoteEmail("tr", companySummary, quote.Number, totalStr, doc, publicLink)
	emailBody.To = cari.Email

	// Send email using the superadmin's configured SMTP (DB settings), not the env/log fallback.
	if err := SendEmail(emailBody); err != nil {
		return errors.New("email gönderilemedi: " + err.Error())
	}

	now := time.Now()
	if quote.SentAt == nil {
		quote.SentAt = &now
	}
	quote.LastSentAt = &now
	quote.Status = "sent"

	if err := txTenant.Save(quote).Error; err != nil {
		return err
	}

	return WriteAuditLog(ctx, txTenant, "quote", quote.ID, "send", userID, "Teklif gönderildi: "+quote.Number)
}

// AcceptPublic ve RejectPublic, public token erişimiyle (tenant context'i olmadan)
// çağrılır; RLS uygulanan database.DB satırı gizler, bu yüzden SystemDB kullanılır.
// Token zaten gizli/tekil olduğundan bu güvenlidir.
func (s *QuoteService) AcceptPublic(ctx context.Context, token string) (*models.Invoice, error) {
	var invoiceRes *models.Invoice
	err := database.SystemDB.Transaction(func(tx *gorm.DB) error {
		var quote models.Quote
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&quote, "public_token = ?", token).Error; err != nil {
			return err
		}

		if quote.Status != "sent" {
			return errors.New("bu teklif daha önce yanıtlanmış veya süresi dolmuş")
		}

		now := time.Now()
		quote.Status = "accepted"
		quote.RespondedAt = &now
		if err := tx.Save(&quote).Error; err != nil {
			return err
		}

		createdBy := uuid.Nil
		if quote.CreatedBy != nil {
			createdBy = *quote.CreatedBy
		}

		// Call internal convert logic inside the same transaction
		// We pass the tx into ctx for the internal function to use
		ctxWithTx := context.WithValue(ctx, utils.TxKey, tx)
		
		var err error
		invoiceRes, err = s.Convert(ctxWithTx, quote.ID, createdBy)
		if err != nil {
			return err
		}

		notifSvc := NewNotificationService()
		title := "Teklif Kabul Edildi: " + quote.Number
		message := "Müşteri teklifi kabul etti. Otomatik fatura oluşturuldu: " + invoiceRes.Number
		if err := notifSvc.CreateNotification(ctxWithTx, tx, quote.CompanyID, "quote_accepted", title, message, &quote.ID, "quote"); err != nil {
			return err
		}

		return WriteAuditLog(ctxWithTx, tx, "quote", quote.ID, "accept_public", createdBy, "Müşteri teklifi kabul etti")
	})

	return invoiceRes, err
}

func (s *QuoteService) RejectPublic(ctx context.Context, token string, note string) error {
	return database.SystemDB.Transaction(func(tx *gorm.DB) error {
		var quote models.Quote
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&quote, "public_token = ?", token).Error; err != nil {
			return err
		}

		if quote.Status != "sent" {
			return errors.New("bu teklif daha önce yanıtlanmış veya süresi dolmuş")
		}

		now := time.Now()
		quote.Status = "rejected"
		quote.RejectNote = note
		quote.RespondedAt = &now
		if err := tx.Save(&quote).Error; err != nil {
			return err
		}

		createdBy := uuid.Nil
		if quote.CreatedBy != nil {
			createdBy = *quote.CreatedBy
		}

		notifSvc := NewNotificationService()
		title := "Teklif Reddedildi: " + quote.Number
		message := "Müşteri teklifi reddetti. Not: " + note
		ctxWithTx := context.WithValue(ctx, utils.TxKey, tx)
		if err := notifSvc.CreateNotification(ctxWithTx, tx, quote.CompanyID, "quote_rejected", title, message, &quote.ID, "quote"); err != nil {
			return err
		}

		return WriteAuditLog(ctx, tx, "quote", quote.ID, "reject_public", createdBy, "Müşteri teklifi reddetti. Not: "+note)
	})
}
