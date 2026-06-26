package services

import (
	"github.com/shopspring/decimal"

	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"os"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)

var (
	ErrInvoiceNotFound        = errors.New("invoice_not_found")
	ErrInvoiceNotEditable     = errors.New("invoice_not_editable")  // Once finalized, cannot edit
	ErrInvoiceNotDeletable    = errors.New("invoice_not_deletable") // Once finalized, cannot delete
	ErrInvoiceAlreadyCanceled = errors.New("invoice_already_canceled")
	ErrCariNotFoundForInvoice = errors.New("cari_not_found")
)

var allowedTransitions = map[string][]string{
	"draft":    {"sent", "canceled"},
	"sent":     {"disputed", "partial", "paid", "canceled"},
	"disputed": {"sent", "canceled"},
	"partial":  {"sent", "paid", "canceled"}, // sent: tam iade olursa
	"paid":     {"sent", "canceled"},          // sent: tam iade olursa
	"canceled": {},
}

func canTransition(from, to string) bool {
	for _, s := range allowedTransitions[from] {
		if s == to {
			return true
		}
	}
	return from == to
}

type InvoiceService struct{}

func NewInvoiceService() *InvoiceService {
	return &InvoiceService{}
}

type InvoiceItemInput struct {
	ProductID      *uuid.UUID      `json:"product_id"`
	Description    string          `json:"description" binding:"max=2000"`
	Quantity       decimal.Decimal `json:"quantity" binding:"required"`
	Unit           string          `json:"unit" binding:"max=50"`
	UnitPrice      decimal.Decimal `json:"unit_price" binding:"required"`
	DiscountRate   decimal.Decimal `json:"discount_rate"`
	TaxRate        decimal.Decimal `json:"tax_rate"`
	Currency       string          `json:"currency" binding:"max=10"`
	ExchangeRate   decimal.Decimal `json:"exchange_rate"`
	ExchangeRateOp string          `json:"exchange_rate_op" binding:"max=5"`
}

// convertToDefaultCurrency, bir satır tutarını kendi dövizinden şirketin
// varsayılan dövizine, satırın kendi kuru ve işlem işaretiyle çevirir.
func convertToDefaultCurrency(amount decimal.Decimal, itemCurrency, defaultCurrency string, rate decimal.Decimal, op string) decimal.Decimal {
	if itemCurrency == "" || itemCurrency == defaultCurrency || rate.IsZero() {
		return amount
	}
	if op == "/" {
		return amount.Div(rate)
	}
	return amount.Mul(rate)
}

type InvoiceInput struct {
	CariID       uuid.UUID          `json:"cari_id" binding:"required"`
	Type         string             `json:"type" binding:"required,oneof=sales purchase"`
	Number       string             `json:"number" binding:"max=100"`
	Date         time.Time          `json:"date" binding:"required"`
	DueDate      time.Time          `json:"due_date" binding:"required"`
	Currency     string             `json:"currency" binding:"max=10"`
	ExchangeRate decimal.Decimal    `json:"exchange_rate"`
	Note         string             `json:"note" binding:"max=2000"`
	Status       string             `json:"status" binding:"required,oneof=draft sent partial paid canceled"`
	// ProcessStock nil/true -> stok işlenir (varsayılan); false -> stok hareketi yaratılmaz.
	ProcessStock *bool              `json:"process_stock"`
	Items        []InvoiceItemInput `json:"items" binding:"required,min=1"`
}

func (s *InvoiceService) Create(ctx context.Context, in InvoiceInput, createdBy uuid.UUID) (*models.Invoice, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, err
	}

	var invoice models.Invoice

	// Run within Transaction
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Enforce tenant connection
		txTenant := utils.GetDB(ctx, tx)

		// 1. Verify Cari exists
		var cari models.Cari
		if err := txTenant.First(&cari, "id = ?", in.CariID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCariNotFoundForInvoice
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
			var key, prefix, settingKey string
			if in.Type == "sales" {
				key = "invoice_sales"
				prefix = "INV-S"
				settingKey = "invoice_sales_prefix"   // satış faturası için ayrı prefix ayarı
			} else {
				key = "invoice_purchase"
				prefix = "BILL"
				settingKey = "invoice_purchase_prefix" // alış faturası için ayrı prefix ayarı
			}
			// Çakışmaya dayanıklı üretim: sayaç ile gerçek veri arasında bir
			// tutarsızlık varsa, boş bir numara bulana kadar sayacı atlatarak ilerle.
			for attempt := 0; attempt < 1000; attempt++ {
				generated, err := utils.GenerateNumberWithSetting(txTenant, companyID, key, settingKey, prefix)
				if err != nil {
					return err
				}
				var count int64
				if err := txTenant.Model(&models.Invoice{}).Where("company_id = ? AND number = ?", companyID, generated).Count(&count).Error; err != nil {
					return err
				}
				if count == 0 {
					number = generated
					break
				}
			}
			if number == "" {
				return errors.New("benzersiz fatura numarası üretilemedi")
			}
		} else {
			// Verify number uniqueness
			var count int64
			if err := txTenant.Model(&models.Invoice{}).Where("company_id = ? AND number = ?", companyID, number).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return errors.New("duplicate_invoice_number")
			}
		}

		// 3. Calculate Totals
		var subtotal, discountTotal, taxTotal, total decimal.Decimal
		invoiceItems := make([]models.InvoiceItem, len(in.Items))

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

			// Invoice-seviyesi toplamlar her zaman şirketin varsayılan dövizinde:
			// her satır kendi kuru/işaretiyle çevrilip toplanır.
			subtotal = subtotal.Add(convertToDefaultCurrency(lineSub, itemCurrency, defaultCurrency, itemRate, itemOp))
			discountTotal = discountTotal.Add(convertToDefaultCurrency(lineDisc, itemCurrency, defaultCurrency, itemRate, itemOp))
			taxTotal = taxTotal.Add(convertToDefaultCurrency(lineTax, itemCurrency, defaultCurrency, itemRate, itemOp))
			total = total.Add(convertToDefaultCurrency(lineTot, itemCurrency, defaultCurrency, itemRate, itemOp))

			itemID, _ := uuid.NewV7()
			invoiceItems[i] = models.InvoiceItem{
				ID:             itemID,
				CompanyID:      companyID,
				ProductID:      item.ProductID,
				Description:    item.Description,
				Quantity:       item.Quantity,
				Unit:           item.Unit,
				UnitPrice:      item.UnitPrice,
				DiscountRate:   item.DiscountRate,
				TaxRate:        item.TaxRate,
				LineTotal:      lineTot, // satırın kendi dövizinde
				Currency:       itemCurrency,
				ExchangeRate:   itemRate,
				ExchangeRateOp: itemOp,
			}
		}

		// İstenen statü 'sent' olsa da fatura önce 'draft' olarak oluşturulur;
		// mail başarıyla gittiğinde sendInvoiceEmailTx statüyü 'sent'e taşır.
		// Mail başarısız olursa transaction rollback olur, ledger/stok hareketi
		// hiç işlenmemiş olur — "sent" hep mailin gerçekten gittiğini garanti eder.
		requestedStatus := in.Status

		invoiceID, _ := uuid.NewV7()
		invoice = models.Invoice{
			ID:            invoiceID,
			CompanyID:     companyID,
			CariID:        in.CariID,
			Type:          in.Type,
			Number:        number,
			Date:          in.Date,
			DueDate:       in.DueDate,
			Currency:      defaultCurrency,
			ExchangeRate:  decimal.NewFromInt(1),
			Subtotal:      subtotal,
			DiscountTotal: discountTotal,
			TaxTotal:      taxTotal,
			Total:         total,
			PaidTotal:     decimal.Zero,
			Status:        "draft",
			ProcessStock:  in.ProcessStock == nil || *in.ProcessStock, // nil -> varsayılan true
			Note:          in.Note,
			PublicToken:   uuid.NewString(), // Added PublicToken
			Items:         invoiceItems,
			CreatedBy:     &createdBy,
		}

		// Save Invoice
		if err := txTenant.Create(&invoice).Error; err != nil {
			return err
		}

		if requestedStatus == "sent" {
			if err := sendInvoiceEmailTx(ctx, txTenant, &invoice, createdBy); err != nil {
				return err
			}
		} else if requestedStatus != "draft" {
			// canTransition draft'tan sadece sent/canceled'a izin verir; create
			// sırasında doğrudan partial/paid gibi bir statü istenmiş olamaz.
			return fmt.Errorf("statü '%s' ile fatura oluşturulamaz", requestedStatus)
		}

		// 4. Ledger Hooks: Post Cari Transaction if NOT draft/canceled
		if invoice.Status != "draft" && invoice.Status != "canceled" {
			if err := s.postLedgerEntry(txTenant, &invoice, &cari, createdBy); err != nil {
				return err
			}
			if err := s.postStockMovements(txTenant, &invoice, createdBy); err != nil {
				return err
			}
		}

		if err := WriteAuditLog(ctx, txTenant, "invoice", invoice.ID, "create", createdBy, invoice.Number); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (s *InvoiceService) Update(ctx context.Context, id uuid.UUID, in InvoiceInput, updatedBy uuid.UUID) (*models.Invoice, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, err
	}

	var invoice models.Invoice

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		// 1. Fetch current invoice with Items
		if err := txTenant.Preload("Items").First(&invoice, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInvoiceNotFound
			}
			return err
		}

		// 2. Only allow editing draft invoices
		if invoice.Status != "draft" {
			return ErrInvoiceNotEditable
		}

		// 3. Verify Cari exists
		var cari models.Cari
		if err := txTenant.First(&cari, "id = ?", in.CariID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCariNotFoundForInvoice
			}
			return err
		}

		// 3b. Fetch company's default currency (multi-currency line conversion anchor)
		var company models.Company
		if err := database.SystemDB.First(&company, "id = ?", companyID).Error; err != nil {
			return err
		}
		defaultCurrency := company.Currency
		if defaultCurrency == "" {
			defaultCurrency = "TRY"
		}

		// 4. Recalculate totals and clear old items
		if err := txTenant.Where("invoice_id = ?", id).Delete(&models.InvoiceItem{}).Error; err != nil {
			return err
		}

		var subtotal, discountTotal, taxTotal, total decimal.Decimal
		invoiceItems := make([]models.InvoiceItem, len(in.Items))

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
			invoiceItems[i] = models.InvoiceItem{
				ID:             itemID,
				CompanyID:      companyID,
				InvoiceID:      id,
				ProductID:      item.ProductID,
				Description:    item.Description,
				Quantity:       item.Quantity,
				Unit:           item.Unit,
				UnitPrice:      item.UnitPrice,
				DiscountRate:   item.DiscountRate,
				TaxRate:        item.TaxRate,
				LineTotal:      lineTot,
				Currency:       itemCurrency,
				ExchangeRate:   itemRate,
				ExchangeRateOp: itemOp,
			}
		}

		requestedStatus := in.Status

		invoice.CariID = in.CariID
		invoice.Date = in.Date
		invoice.DueDate = in.DueDate
		invoice.Note = in.Note
		invoice.ProcessStock = in.ProcessStock == nil || *in.ProcessStock
		invoice.Status = "draft"
		invoice.Subtotal = subtotal
		invoice.DiscountTotal = discountTotal
		invoice.TaxTotal = taxTotal
		invoice.Total = total
		invoice.Items = invoiceItems
		invoice.UpdatedBy = &updatedBy
		invoice.Currency = defaultCurrency
		invoice.ExchangeRate = decimal.NewFromInt(1)

		// Save Invoice and new Items
		if err := txTenant.Save(&invoice).Error; err != nil {
			return err
		}

		if requestedStatus == "sent" {
			if err := sendInvoiceEmailTx(ctx, txTenant, &invoice, updatedBy); err != nil {
				return err
			}
		} else if requestedStatus != "draft" {
			return fmt.Errorf("statü '%s' ile fatura güncellenemez", requestedStatus)
		}

		// 5. Ledger Hooks: Post Cari Transaction if status has changed to final
		if invoice.Status != "draft" && invoice.Status != "canceled" {
			createdByVal := uuid.Nil
			if invoice.CreatedBy != nil {
				createdByVal = *invoice.CreatedBy
			}
			if err := s.postLedgerEntry(txTenant, &invoice, &cari, createdByVal); err != nil {
				return err
			}
			if err := s.postStockMovements(txTenant, &invoice, createdByVal); err != nil {
				return err
			}
		}

		if err := WriteAuditLog(ctx, txTenant, "invoice", invoice.ID, "update", updatedBy, invoice.Number); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (s *InvoiceService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		var invoice models.Invoice
		if err := txTenant.First(&invoice, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInvoiceNotFound
			}
			return err
		}

		// Only allow deleting drafts
		if invoice.Status != "draft" {
			return ErrInvoiceNotDeletable
		}

		if err := WriteAuditLog(ctx, txTenant, "invoice", invoice.ID, "delete", userID, invoice.Number); err != nil {
			return err
		}

		// Items will cascade delete due to DB constraints set in migrate.go / tags
		return txTenant.Delete(&invoice).Error
	})
}

func (s *InvoiceService) GetByID(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	tx := utils.GetDB(ctx, database.DB)

	var invoice models.Invoice
	if err := tx.Preload("Items").Preload("CreatedByUser").Preload("UpdatedByUser").First(&invoice, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvoiceNotFound
		}
		return nil, err
	}
	return &invoice, nil
}

func (s *InvoiceService) List(ctx context.Context, page, limit int, query, sort string, filters map[string]string) ([]models.Invoice, int64, error) {
	tx := utils.GetDB(ctx, database.DB)

	var invoices []models.Invoice
	var total int64

	dbQuery := tx.Model(&models.Invoice{})

	if query != "" {
		q := "%" + query + "%"
		dbQuery = dbQuery.Where("number ILIKE ? OR note ILIKE ?", q, q)
	}

	if typeFilter, exists := filters["type"]; exists && typeFilter != "" {
		dbQuery = dbQuery.Where("type = ?", typeFilter)
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
	if err := dbQuery.Preload("CreatedByUser").Preload("UpdatedByUser").Offset(offset).Limit(limit).Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

func (s *InvoiceService) Cancel(ctx context.Context, id uuid.UUID, createdBy uuid.UUID) error {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return err
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		// 1. Fetch invoice
		var invoice models.Invoice
		if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&invoice, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInvoiceNotFound
			}
			return err
		}

		if invoice.Status == "canceled" {
			return ErrInvoiceAlreadyCanceled
		}

		oldStatus := invoice.Status
		invoice.Status = "canceled"

		// Save status change
		if err := txTenant.Model(&invoice).Update("status", "canceled").Error; err != nil {
			return err
		}

		// If it was a draft, no ledger entries exist, so no reversing is needed.
		if oldStatus == "draft" {
			if err := WriteAuditLog(ctx, txTenant, "invoice", invoice.ID, "cancel", createdBy, invoice.Number); err != nil {
				return err
			}
			return nil
		}

		// 2. Post Reversing Ledger Entry
		var cari models.Cari
		if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&cari, "id = ? AND company_id = ?", invoice.CariID, companyID).Error; err != nil {
			return err
		}

		// Reversing entry logic:
		// If original was Sales (debit), we write Credit (reversing).
		// If original was Purchase (credit), we write Debit (reversing).
		var txType string
		var amountDelta decimal.Decimal

		if invoice.Type == "sales" {
			txType = "credit"
			amountDelta = invoice.Total.Neg()
		} else {
			txType = "debit"
			amountDelta = invoice.Total
		}

		balanceAfter, err := UpdateCariBalance(txTenant, cari.ID, invoice.Currency, amountDelta)
		if err != nil {
			return err
		}

		txID, _ := uuid.NewV7()
		reversingTx := models.CariTransaction{
			ID:           txID,
			CompanyID:    companyID,
			CariID:       invoice.CariID,
			Date:         utils.NowIn(ctx),
			Type:         txType,
			SourceType:   "invoice",
			Currency:     invoice.Currency,
			SourceID:     &invoice.ID,
			Description:  fmt.Sprintf("İptal - Fatura No: %s", invoice.Number),
			Amount:       invoice.Total,
			BalanceAfter: balanceAfter,
			CreatedBy:    &createdBy,
		}

		if err := txTenant.Create(&reversingTx).Error; err != nil {
			return err
		}

		// Update cached balance inside UpdateCariBalance
		// We no longer update cari.Balance here

		// 3. Post Reversing Stock Movements
		if err := s.revertStockMovements(ctx, txTenant, &invoice, createdBy); err != nil {
			return err
		}

		if err := WriteAuditLog(ctx, txTenant, "invoice", invoice.ID, "cancel", createdBy, invoice.Number); err != nil {
			return err
		}

		return nil
	})
}

// UpdateStatus allows changing the status (and paid_total) of a non-draft invoice,
// e.g. marking a 'sent' invoice as 'paid'/'partial' manually from the UI.
// Unlike Update, this does not touch items/totals and is not restricted to drafts.
func (s *InvoiceService) UpdateStatus(ctx context.Context, id uuid.UUID, status string, paidTotal decimal.Decimal, updatedBy uuid.UUID) (*models.Invoice, error) {
	var invoice models.Invoice

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&invoice, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInvoiceNotFound
			}
			return err
		}

		if !canTransition(invoice.Status, status) {
			return fmt.Errorf("invalid_status_transition: %s -> %s", invoice.Status, status)
		}

		invoice.Status = status
		invoice.PaidTotal = paidTotal
		invoice.UpdatedBy = &updatedBy

		if err := txTenant.Model(&invoice).Updates(map[string]interface{}{
			"status":     status,
			"paid_total": paidTotal,
			"updated_by": updatedBy,
		}).Error; err != nil {
			return err
		}

		if err := WriteAuditLog(ctx, txTenant, "invoice", invoice.ID, "update_status", updatedBy, invoice.Number+" - "+status); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// postLedgerEntry writes a Cari ledger log and adjusts cached balance.
func (s *InvoiceService) postLedgerEntry(tx *gorm.DB, invoice *models.Invoice, cari *models.Cari, createdBy uuid.UUID) error {
	// Idempotency guard: skip if a ledger entry for this invoice already exists
	var existing int64
	if err := tx.Model(&models.CariTransaction{}).
		Where("source_type = ? AND source_id = ?", "invoice", invoice.ID).
		Count(&existing).Error; err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}

	// Lock Cari row to guarantee balance safety
	var lockedCari models.Cari
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lockedCari, "id = ?", cari.ID).Error; err != nil {
		return err
	}

	var txType string
	var amountDelta decimal.Decimal

	// Sales invoice: Cari is debited (+ balance goes up)
	// Purchase invoice: Cari is credited (- balance goes down)
	if invoice.Type == "sales" {
		txType = "debit"
		amountDelta = invoice.Total
	} else {
		txType = "credit"
		amountDelta = invoice.Total.Neg()
	}

	balanceAfter, err := UpdateCariBalance(tx, lockedCari.ID, invoice.Currency, amountDelta)
	if err != nil {
		return err
	}

	txID, _ := uuid.NewV7()
	ledgerTx := models.CariTransaction{
		ID:           txID,
		CompanyID:    invoice.CompanyID,
		CariID:       invoice.CariID,
		Date:         invoice.Date,
		Type:         txType,
		SourceType:   "invoice",
		SourceID:     &invoice.ID,
		Currency:     invoice.Currency,
		Description:  fmt.Sprintf("%s Faturası No: %s", getInvoiceTypeLabel(invoice.Type), invoice.Number),
		Amount:       invoice.Total,
		BalanceAfter: balanceAfter,
		CreatedBy:    &createdBy,
	}

	if err := tx.Create(&ledgerTx).Error; err != nil {
		return err
	}

	// Cached balance already updated in UpdateCariBalance
	return nil
}

func getInvoiceTypeLabel(t string) string {
	if t == "sales" {
		return "Satış"
	}
	return "Alış"
}

// postStockMovements writes stock movements for non-draft finalized invoice items.
func (s *InvoiceService) postStockMovements(tx *gorm.DB, invoice *models.Invoice, createdBy uuid.UUID) error {
	// Kullanıcı bu fatura için stok işlemeyi kapattıysa hiç hareket yaratma
	// (hizmet/gider faturaları gibi).
	if !invoice.ProcessStock {
		return nil
	}

	// Idempotency guard: skip if stock movements for this invoice already exist
	var existingMov int64
	if err := tx.Model(&models.StockMovement{}).
		Where("source_type = ? AND source_id = ?", "invoice", invoice.ID).
		Count(&existingMov).Error; err != nil {
		return err
	}
	if existingMov > 0 {
		return nil
	}

	stockSvc := NewStockService()
	whID, err := stockSvc.GetOrCreateDefaultWarehouse(tx, invoice.CompanyID, createdBy)
	if err != nil {
		return err
	}

	for _, item := range invoice.Items {
		if item.ProductID == nil {
			continue
		}

		var prod models.Product
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&prod, "id = ?", *item.ProductID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}

		if !prod.TrackStock {
			continue
		}

		qty := item.Quantity
		// Stok maliyeti her zaman base (şirket varsayılan) dövizinde tutulur.
		// invoice.Currency base'dir; satır farklı dövizdeyse kendi donmuş kuruyla
		// base'e çevrilir (USD/EUR/RUB alış -> TL maliyet). Aksi halde farklı
		// dövizdeki alışlar ortalama maliyeti bozardı.
		baseUnitPrice := convertToDefaultCurrency(item.UnitPrice, item.Currency, invoice.Currency, item.ExchangeRate, item.ExchangeRateOp)
		var mType string
		var newStock decimal.Decimal
		var newAverageCost = prod.AverageCost // Varsayılan olarak değişmez

		if invoice.Type == "sales" {
			mType = "out"
			newStock = prod.CurrentStock.Sub(qty)
			if newStock.LessThan(decimal.Zero) {
				return fmt.Errorf("stok yetersiz: %s ürünü için kalan stok %s, istenen %s", prod.Name, prod.CurrentStock.String(), qty.String())
			}
		} else { // purchase (alış faturası = stok girişi)
			mType = "in"
			newStock = prod.CurrentStock.Add(qty)

			// WAC (Ağırlıklı Ortalama Maliyet) Hesaplaması — base dövizinde
			if newStock.GreaterThan(decimal.Zero) {
				totalOldValue := prod.CurrentStock.Mul(prod.AverageCost)
				discountFactor := decimal.NewFromInt(1).Sub(item.DiscountRate.Div(decimal.NewFromInt(100)))
				totalNewValue := qty.Mul(baseUnitPrice).Mul(discountFactor)
				newAverageCost = totalOldValue.Add(totalNewValue).DivRound(newStock, 4)
			}
		}

		movementID, _ := uuid.NewV7()
		movement := models.StockMovement{
			ID:           movementID,
			CompanyID:    invoice.CompanyID,
			ProductID:    *item.ProductID,
			WarehouseID:  whID,
			Date:         invoice.Date,
			Type:         mType,
			SourceType:   "invoice",
			SourceID:     &invoice.ID,
			Quantity:     item.Quantity,
			UnitCost:     baseUnitPrice, // base dövizinde birim maliyet
			BalanceAfter: newStock,
			Note:         fmt.Sprintf("Fatura No: %s", invoice.Number),
			CreatedBy:    &createdBy,
		}

		if err := tx.Create(&movement).Error; err != nil {
			return err
		}

		if err := tx.Model(&prod).Updates(map[string]interface{}{
			"current_stock": newStock,
			"average_cost":  newAverageCost,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

// revertStockMovements writes reversing stock movements when canceling an invoice.
func (s *InvoiceService) revertStockMovements(ctx context.Context, tx *gorm.DB, invoice *models.Invoice, createdBy uuid.UUID) error {
	// Stok işlemeyen fatura iptal edilirken de ters hareket yaratılmaz; aksi halde
	// hiç giriş yapılmamış stok eksiye/fazlaya kayardı.
	if !invoice.ProcessStock {
		return nil
	}

	var existingRev int64
	if err := tx.Model(&models.StockMovement{}).
		Where("source_type = ? AND source_id = ?", "invoice_cancel", invoice.ID).
		Count(&existingRev).Error; err != nil {
		return err
	}
	if existingRev > 0 {
		return nil
	}

	stockSvc := NewStockService()
	whID, err := stockSvc.GetOrCreateDefaultWarehouse(tx, invoice.CompanyID, createdBy)
	if err != nil {
		return err
	}

	var items []models.InvoiceItem
	if err := tx.Where("invoice_id = ?", invoice.ID).Find(&items).Error; err != nil {
		return err
	}

	for _, item := range items {
		if item.ProductID == nil {
			continue
		}

		var prod models.Product
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&prod, "id = ?", *item.ProductID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}

		if !prod.TrackStock {
			continue
		}

		qty := item.Quantity
		// İptalde de maliyet base dövizinde düşülür (girişteki ile aynı çevrim).
		baseUnitPrice := convertToDefaultCurrency(item.UnitPrice, item.Currency, invoice.Currency, item.ExchangeRate, item.ExchangeRateOp)
		var mType string
		var newStock decimal.Decimal
		var newAverageCost = prod.AverageCost

		if invoice.Type == "sales" { // satış iptali = stok girişi
			mType = "in"
			newStock = prod.CurrentStock.Add(qty)

			// Satış iptalinde maliyetin (average_cost) o günkü fiyattan güncellenip güncellenmeyeceği
			// tartışmalıdır, ancak genellikle stok aynı maliyetle (iade) girdiği için maliyeti değiştirmemek
			// veya çıkış anındaki ortalama maliyeti korumak daha güvenlidir. Bu yüzden burada WAC uygulanmıyor,
			// çünkü satış faturalarında "UnitPrice" satış fiyatıdır, alış maliyeti değil.

		} else { // purchase iptali = stok çıkışı
			mType = "out"
			newStock = prod.CurrentStock.Sub(qty)
			if newStock.LessThan(decimal.Zero) {
				return fmt.Errorf("stok yetersiz: iptal edilmek istenen alis faturasindaki %s ürünü için kalan stok %s, düşülmek istenen %s", prod.Name, prod.CurrentStock.String(), qty.String())
			}

			// WAC düzeltmesi: iptal edilen base değeri toplam değerden çıkar
			if newStock.GreaterThan(decimal.Zero) {
				totalCurrentValue := prod.CurrentStock.Mul(prod.AverageCost)
				discountFactor := decimal.NewFromInt(1).Sub(item.DiscountRate.Div(decimal.NewFromInt(100)))
				cancelledValue := qty.Mul(baseUnitPrice).Mul(discountFactor)
				newAverageCost = totalCurrentValue.Sub(cancelledValue).DivRound(newStock, 4)
			}
		}

		movementID, _ := uuid.NewV7()
		movement := models.StockMovement{
			ID:           movementID,
			CompanyID:    invoice.CompanyID,
			ProductID:    *item.ProductID,
			WarehouseID:  whID,
			Date:         utils.NowIn(ctx),
			Type:         mType,
			SourceType:   "invoice_cancel",
			SourceID:     &invoice.ID,
			Quantity:     item.Quantity,
			UnitCost:     baseUnitPrice,
			BalanceAfter: newStock,
			Note:         fmt.Sprintf("İptal - Fatura No: %s", invoice.Number),
			CreatedBy:    &createdBy,
		}

		if err := tx.Create(&movement).Error; err != nil {
			return err
		}

		if err := tx.Model(&prod).Updates(map[string]interface{}{
			"current_stock": newStock,
			"average_cost":  newAverageCost,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

// SendInvoice sends the invoice public link via email to the Cari contact.
func (s *InvoiceService) SendInvoice(ctx context.Context, invoiceID, userID uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		var invoice models.Invoice
		if err := txTenant.First(&invoice, "id = ?", invoiceID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInvoiceNotFound
			}
			return err
		}

		if !canTransition(invoice.Status, "sent") {
			return fmt.Errorf("statü '%s' iken gönderilemez", invoice.Status)
		}

		return sendInvoiceEmailTx(ctx, txTenant, &invoice, userID)
	})
}

// sendInvoiceEmailTx, faturanın mail gönderim+statü güncelleme mantığını taşır.
// Hem SendInvoice (manuel "Gönder" butonu) hem Create/Update (durum doğrudan
// 'sent' seçildiğinde) tarafından, ikisi de kendi transaction'larının içinde
// olacak şekilde çağrılır; mail başarısız olursa hata döner ve çağıran
// transaction'ı rollback eder (status hiç 'sent' yazılmaz).
func sendInvoiceEmailTx(ctx context.Context, txTenant *gorm.DB, invoice *models.Invoice, userID uuid.UUID) error {
	var cari models.Cari
	if err := txTenant.First(&cari, "id = ?", invoice.CariID).Error; err != nil {
		return err
	}

	if cari.Email == "" {
		return errors.New("cari_email_not_found")
	}

	var company models.Company
	if err := database.SystemDB.First(&company, "id = ?", invoice.CompanyID).Error; err != nil {
		return err
	}

	// Ödeme linki yalnızca satış faturalarında anlamlı (müşteri görüntüleyip öder).
	// Alış faturasında tedarikçiye public ödeme sayfası sunmak mantıksız olduğundan
	// token/link üretimini sadece sales'te yapıyoruz.
	link := ""
	if invoice.Type == "sales" {
		// Eski kayıtlarda (PublicToken alanı eklenmeden önce oluşturulmuş
		// faturalarda) bu alan boş olabilir; boş token'la link üretip kırık
		// bir mail göndermek yerine burada token'ı tamamlayıp devam ediyoruz.
		if invoice.PublicToken == "" {
			invoice.PublicToken = uuid.NewString()
			if err := txTenant.Model(invoice).Update("public_token", invoice.PublicToken).Error; err != nil {
				return err
			}
		}

		baseURL := os.Getenv("PUBLIC_APP_URL")
		if baseURL == "" {
			baseURL = "http://localhost:5173" // fallback
		}
		link = fmt.Sprintf("%s/pay/%s", baseURL, invoice.PublicToken)
	}

	companySummary := utils.CompanySummary{
		Name:    company.Name,
		Phone:   company.Phone,
		Email:   company.Email,
		Address: company.Address,
	}

	lines := make([]utils.DocumentLine, 0, len(invoice.Items))
	for _, item := range invoice.Items {
		// Satır tutarları satırın KENDİ dövizinde; bu yüzden satır dövizi ile
		// etiketlenir (önceden base etiketi yapıştırılıp yanlış görünüyordu).
		lineCur := item.Currency
		if lineCur == "" {
			lineCur = invoice.Currency
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
	if invoice.DiscountTotal.IsPositive() {
		discountText = "-" + invoice.DiscountTotal.StringFixed(2) + " " + invoice.Currency
	}
	doc := utils.DocumentSummary{
		Number:       invoice.Number,
		Date:         invoice.Date.Format("02.01.2006"),
		DueDate:      invoice.DueDate.Format("02.01.2006"),
		Customer:     utils.DocumentParty{Name: custName, Address: cari.Address},
		Lines:        lines,
		Subtotal:     invoice.Subtotal.StringFixed(2) + " " + invoice.Currency,
		DiscountText: discountText,
		TaxText:      invoice.TaxTotal.StringFixed(2) + " " + invoice.Currency,
		Total:        invoice.Total.StringFixed(2) + " " + invoice.Currency,
		Currency:     invoice.Currency,
	}

	amountStr := invoice.Total.StringFixed(2) + " " + invoice.Currency
	var emailReq utils.Email
	if invoice.Type == "purchase" {
		// Alış: tedarikçiye "faturanız tarafımızca alındı/kaydedildi" teyidi (ödeme linki YOK).
		emailReq = utils.PurchaseInvoiceEmail("tr", companySummary, invoice.Number, amountStr, doc)
	} else {
		// Satış: müşteriye "faturanız kesildi, görüntüleyin ve ödeyin" + ödeme linki.
		emailReq = utils.InvoiceEmail("tr", companySummary, invoice.Number, amountStr, doc, link)
	}
	emailReq.To = cari.Email

	// Send email using the superadmin's configured SMTP (DB settings), not the env/log fallback.
	if err := SendEmail(emailReq); err != nil {
		return fmt.Errorf("email_send_failed: %w", err)
	}

	now := utils.NowIn(ctx)
	updates := map[string]interface{}{
		"status":       "sent",
		"last_sent_at": now,
	}
	if invoice.SentAt == nil {
		updates["sent_at"] = now
	}

	if err := txTenant.Model(invoice).Updates(updates).Error; err != nil {
		return err
	}

	return WriteAuditLog(ctx, txTenant, "invoice", invoice.ID, "send", userID, invoice.Number)
}
