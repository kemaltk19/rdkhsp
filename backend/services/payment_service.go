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

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)

var (
	ErrPaymentNotFound        = errors.New("payment_not_found")
	ErrPaymentAlreadyCanceled = errors.New("payment_already_canceled")
	ErrCariNotFoundForPayment = errors.New("cari_not_found")
	ErrAccountNotFound        = errors.New("account_not_found")
	ErrCurrencyMismatch       = errors.New("currency_mismatch")
	ErrInvoiceNotFoundForPay  = errors.New("invoice_not_found")
)

type PaymentService struct{}

func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

type PaymentInput struct {
	CariID      uuid.UUID       `json:"cari_id" binding:"required"`
	Type        string          `json:"type" binding:"required,oneof=collection payment"`
	Date        time.Time       `json:"date" binding:"required"`
	Method      string          `json:"method" binding:"required,oneof=cash bank card check"`
	AccountKind string          `json:"account_kind" binding:"required,oneof=cash bank"`
	AccountID   uuid.UUID       `json:"account_id" binding:"required"`
	Amount      decimal.Decimal `json:"amount" binding:"required"`
	Currency    string          `json:"currency" binding:"max=10"`
	InvoiceID   *uuid.UUID      `json:"invoice_id"`
	Reference   string          `json:"reference" binding:"max=100"`
	Note        string          `json:"note" binding:"max=2000"`
}

func (s *PaymentService) Create(ctx context.Context, in PaymentInput, createdBy uuid.UUID) (*models.Payment, error) {
	if in.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("ödeme tutarı sıfır veya sıfırdan küçük olamaz")
	}

	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, err
	}

	var payment models.Payment

	runInTx := func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		// 1. Verify Cari
		var cari models.Cari
		if err := txTenant.First(&cari, "id = ?", in.CariID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCariNotFoundForPayment
			}
			return err
		}

		// 2. Önce hesabın dövizini al: ödeme fiziksel olarak hesabın dövizinde
		// hareket eder, bu yüzden ödeme dövizi varsayılanı hesabın dövizidir
		// (eski hali cari.Currency idi -> base cari borcuyla çelişebiliyordu).
		var accountCurrency string
		if in.AccountKind == "cash" {
			var acc models.CashAccount
			if err := txTenant.First(&acc, "id = ?", in.AccountID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrAccountNotFound
				}
				return err
			}
			accountCurrency = acc.Currency
		} else {
			var acc models.BankAccount
			if err := txTenant.First(&acc, "id = ?", in.AccountID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrAccountNotFound
				}
				return err
			}
			accountCurrency = acc.Currency
		}

		currency := in.Currency
		if currency == "" {
			currency = accountCurrency
		}
		if accountCurrency != currency {
			return ErrCurrencyMismatch
		}

		// 3. Verify Invoice if provided
		var invoice models.Invoice
		if in.InvoiceID != nil {
			if err := txTenant.First(&invoice, "id = ?", *in.InvoiceID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrInvoiceNotFoundForPay
				}
				return err
			}
			if invoice.Currency != currency {
				return ErrCurrencyMismatch
			}
		}

		paymentID, _ := uuid.NewV7()
		
		reference := in.Reference
		if reference == "" {
			settingKey := "collection_prefix"
			seqKey := "collection"
			defaultPrefix := "TAH"
			if in.Type == "payment" {
				settingKey = "payment_prefix"
				seqKey = "payment_out"
				defaultPrefix = "ODE"
			}
			reference, _ = utils.GenerateNumberWithSetting(txTenant, companyID, seqKey, settingKey, defaultPrefix)
		}

		payment = models.Payment{
			ID:          paymentID,
			CompanyID:   companyID,
			CariID:      in.CariID,
			Type:        in.Type,
			Date:        in.Date,
			Method:      in.Method,
			AccountKind: in.AccountKind,
			AccountID:   in.AccountID,
			Amount:      in.Amount,
			Currency:    currency,
			InvoiceID:   in.InvoiceID,
			Reference:   reference,
			Note:        in.Note,
			Status:      "completed",
			CreatedBy:   &createdBy,
		}

		if err := txTenant.Create(&payment).Error; err != nil {
			return err
		}

		// 4. Ledger Hook: Cari Transaction
		var lockedCari models.Cari
		if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lockedCari, "id = ?", cari.ID).Error; err != nil {
			return err
		}

		var cariTxType string
		var amountDelta decimal.Decimal

		// Collection (tahsilat): customer pays us -> credit Cari (- balance goes down)
		// Payment (ödeme): we pay supplier -> debit Cari (+ balance goes up)
		if in.Type == "collection" {
			cariTxType = "credit"
			amountDelta = in.Amount.Neg()
		} else {
			cariTxType = "debit"
			amountDelta = in.Amount
		}

		cariBalanceAfter, err := UpdateCariBalance(txTenant, lockedCari.ID, payment.Currency, amountDelta)
		if err != nil {
			return err
		}

		cariTxID, _ := uuid.NewV7()
		cariTx := models.CariTransaction{
			ID:           cariTxID,
			CompanyID:    companyID,
			CariID:       in.CariID,
			Date:         in.Date,
			Type:         cariTxType,
			SourceType:   "payment",
			SourceID:     &payment.ID,
			Currency:     payment.Currency,
			Description:  fmt.Sprintf("%s (%s) - Ref: %s", getPaymentTypeLabel(in.Type), getMethodLabel(in.Method), in.Reference),
			Amount:       in.Amount,
			BalanceAfter: cariBalanceAfter,
			CreatedBy:    &createdBy,
		}
		if err := txTenant.Create(&cariTx).Error; err != nil {
			return err
		}
		// Cached balance updated via UpdateCariBalance

		// 5. Ledger Hook: Cash/Bank Transaction
		var accountBalance decimal.Decimal
		if in.AccountKind == "cash" {
			var acc models.CashAccount
			if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&acc, "id = ?", in.AccountID).Error; err != nil {
				return err
			}
			accountBalance = acc.Balance
		} else {
			var acc models.BankAccount
			if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&acc, "id = ?", in.AccountID).Error; err != nil {
				return err
			}
			accountBalance = acc.Balance
		}

		var cashTxType string
		var cashBalanceAfter decimal.Decimal

		// Collection: money comes in -> "in" (+ balance goes up)
		// Payment: money goes out -> "out" (- balance goes down)
		if in.Type == "collection" {
			cashTxType = "in"
			cashBalanceAfter = accountBalance.Add(in.Amount)
		} else {
			cashTxType = "out"
			cashBalanceAfter = accountBalance.Sub(in.Amount)
		}

		cashTxID, _ := uuid.NewV7()
		cashTx := models.CashTransaction{
			ID:           cashTxID,
			CompanyID:    companyID,
			AccountKind:  in.AccountKind,
			AccountID:    in.AccountID,
			Date:         in.Date,
			Type:         cashTxType,
			SourceType:   "payment",
			SourceID:     &payment.ID,
			Amount:       in.Amount,
			BalanceAfter: cashBalanceAfter,
			Description:  fmt.Sprintf("Cari %s: %s", getPaymentTypeLabel(in.Type), lockedCari.Name),
			CreatedBy:    &createdBy,
		}
		if err := txTenant.Create(&cashTx).Error; err != nil {
			return err
		}

		// Update Account balance
		if in.AccountKind == "cash" {
			if err := txTenant.Table("cash_accounts").Where("id = ?", in.AccountID).Update("balance", cashBalanceAfter).Error; err != nil {
				return err
			}
		} else {
			if err := txTenant.Table("bank_accounts").Where("id = ?", in.AccountID).Update("balance", cashBalanceAfter).Error; err != nil {
				return err
			}
		}

		// 6. Invoice Allocation Update
		if in.InvoiceID != nil {
			var lockedInvoice models.Invoice
			if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lockedInvoice, "id = ?", *in.InvoiceID).Error; err != nil {
				return err
			}

			newPaidTotal := lockedInvoice.PaidTotal.Add(in.Amount)
			lockedInvoice.PaidTotal = newPaidTotal

			var newStatus string
			if newPaidTotal.GreaterThanOrEqual(lockedInvoice.Total) {
				newStatus = "paid"
			} else if newPaidTotal.GreaterThan(decimal.Zero) {
				newStatus = "partial"
			} else {
				newStatus = "sent"
			}

			if !canTransition(lockedInvoice.Status, newStatus) {
				return fmt.Errorf("fatura statüsü '%s' iken '%s' durumuna geçişe izin verilmiyor", lockedInvoice.Status, newStatus)
			}
			lockedInvoice.Status = newStatus

			if err := txTenant.Save(&lockedInvoice).Error; err != nil {
				return err
			}
		}

		summaryStr := fmt.Sprintf("%s - %s %s", getPaymentTypeLabel(payment.Type), payment.Amount.StringFixed(2), payment.Currency)
		if err := WriteAuditLog(ctx, txTenant, "payment", payment.ID, "create", createdBy, summaryStr); err != nil {
			return err
		}

		return nil
	}

	// Context'te zaten bir transaction varsa (örn. public Pay() handler'ı
	// satırı kilitleyip çağırıyorsa) onu yeniden kullan; aksi halde yeni aç.
	// Aksi halde dış tx'in tuttuğu satır kilitleri üzerinde içeride yeni bir
	// transaction açmak deadlock riski taşır.
	if existingTx, ok := ctx.Value(utils.TxKey).(*gorm.DB); ok && existingTx != nil {
		err = runInTx(existingTx)
	} else {
		err = database.DB.Transaction(runInTx)
	}

	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (s *PaymentService) Cancel(ctx context.Context, id uuid.UUID, createdBy uuid.UUID) error {
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

		// 1. Get Payment
		var payment models.Payment
		if err := txTenant.First(&payment, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPaymentNotFound
			}
			return err
		}

		if payment.Status == "canceled" {
			return ErrPaymentAlreadyCanceled
		}

		// Update payment status
		payment.Status = "canceled"
		payment.UpdatedBy = &createdBy
		if err := txTenant.Save(&payment).Error; err != nil {
			return err
		}

		// 2. Reversing Cari Transaction
		var cari models.Cari
		if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&cari, "id = ?", payment.CariID).Error; err != nil {
			return err
		}

		var revCariType string
		var revAmountDelta decimal.Decimal

		// Original was collection (credit) -> reversal is debit (+ balance goes up)
		// Original was payment (debit) -> reversal is credit (- balance goes down)
		if payment.Type == "collection" {
			revCariType = "debit"
			revAmountDelta = payment.Amount
		} else {
			revCariType = "credit"
			revAmountDelta = payment.Amount.Neg()
		}

		revCariBalanceAfter, err := UpdateCariBalance(txTenant, cari.ID, payment.Currency, revAmountDelta)
		if err != nil {
			return err
		}

		revCariTxID, _ := uuid.NewV7()
		revCariTx := models.CariTransaction{
			ID:           revCariTxID,
			CompanyID:    companyID,
			CariID:       payment.CariID,
			Date:         utils.NowIn(ctx),
			Type:         revCariType,
			SourceType:   "payment",
			SourceID:     &payment.ID,
			Currency:     payment.Currency,
			Amount:       payment.Amount,
			BalanceAfter: revCariBalanceAfter,
			CreatedBy:    &createdBy,
		}
		if err := txTenant.Create(&revCariTx).Error; err != nil {
			return err
		}
		// Cached balance updated via UpdateCariBalance


		// 3. Reversing Cash/Bank Transaction
		var accountBalance decimal.Decimal
		if payment.AccountKind == "cash" {
			var acc models.CashAccount
			if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&acc, "id = ?", payment.AccountID).Error; err != nil {
				return err
			}
			accountBalance = acc.Balance
		} else {
			var acc models.BankAccount
			if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&acc, "id = ?", payment.AccountID).Error; err != nil {
				return err
			}
			accountBalance = acc.Balance
		}

		var revCashType string
		var revCashBalanceAfter decimal.Decimal

		// Original was collection (in) -> reversal is out (- balance goes down)
		// Original was payment (out) -> reversal is in (+ balance goes up)
		if payment.Type == "collection" {
			revCashType = "out"
			revCashBalanceAfter = accountBalance.Sub(payment.Amount)
		} else {
			revCashType = "in"
			revCashBalanceAfter = accountBalance.Add(payment.Amount)
		}

		revCashTxID, _ := uuid.NewV7()
		revCashTx := models.CashTransaction{
			ID:           revCashTxID,
			CompanyID:    companyID,
			AccountKind:  payment.AccountKind,
			AccountID:    payment.AccountID,
			Date:         utils.NowIn(ctx),
			Type:         revCashType,
			SourceType:   "payment",
			SourceID:     &payment.ID,
			Amount:       payment.Amount,
			BalanceAfter: revCashBalanceAfter,
			Description:  fmt.Sprintf("İptal - Cari %s: %s", getPaymentTypeLabel(payment.Type), cari.Name),
			CreatedBy:    &createdBy,
		}
		if err := txTenant.Create(&revCashTx).Error; err != nil {
			return err
		}

		// Update Account balance
		if payment.AccountKind == "cash" {
			if err := txTenant.Table("cash_accounts").Where("id = ?", payment.AccountID).Update("balance", revCashBalanceAfter).Error; err != nil {
				return err
			}
		} else {
			if err := txTenant.Table("bank_accounts").Where("id = ?", payment.AccountID).Update("balance", revCashBalanceAfter).Error; err != nil {
				return err
			}
		}

		// 4. Reversing Invoice Allocation
		if payment.InvoiceID != nil {
			var lockedInvoice models.Invoice
			if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lockedInvoice, "id = ?", *payment.InvoiceID).Error; err != nil {
				return err
			}

			newPaidTotal := lockedInvoice.PaidTotal.Sub(payment.Amount)
			if newPaidTotal.LessThan(decimal.Zero) {
				newPaidTotal = decimal.Zero
			}
			lockedInvoice.PaidTotal = newPaidTotal

			var newStatus string
			if newPaidTotal.GreaterThanOrEqual(lockedInvoice.Total) {
				newStatus = "paid"
			} else if newPaidTotal.GreaterThan(decimal.Zero) {
				newStatus = "partial"
			} else {
				newStatus = "sent"
			}

			if !canTransition(lockedInvoice.Status, newStatus) {
				return fmt.Errorf("fatura statüsü '%s' iken '%s' durumuna geçişe izin verilmiyor", lockedInvoice.Status, newStatus)
			}
			lockedInvoice.Status = newStatus

			if err := txTenant.Save(&lockedInvoice).Error; err != nil {
				return err
			}
		}

		summaryStr := fmt.Sprintf("%s - %s %s", getPaymentTypeLabel(payment.Type), payment.Amount.StringFixed(2), payment.Currency)
		if err := WriteAuditLog(ctx, txTenant, "payment", payment.ID, "cancel", createdBy, summaryStr); err != nil {
			return err
		}

		return nil
	})
}

func (s *PaymentService) GetByID(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
	tx := utils.GetDB(ctx, database.DB)

	var p models.Payment
	if err := tx.Preload("CreatedByUser").Preload("UpdatedByUser").First(&p, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (s *PaymentService) List(ctx context.Context, page, limit int, query, sort string, filters map[string]string) ([]models.Payment, int64, error) {
	tx := utils.GetDB(ctx, database.DB)

	var payments []models.Payment
	var total int64

	dbQuery := tx.Model(&models.Payment{}).Preload("CreatedByUser").Preload("UpdatedByUser")

	if query != "" {
		q := "%" + strings.ToLower(query) + "%"
		dbQuery = dbQuery.Where("LOWER(reference) LIKE ? OR LOWER(note) LIKE ?", q, q)
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

	if invoiceFilter, exists := filters["invoice_id"]; exists && invoiceFilter != "" {
		dbQuery = dbQuery.Where("invoice_id = ?", invoiceFilter)
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
	if err := dbQuery.Offset(offset).Limit(limit).Find(&payments).Error; err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

func (s *PaymentService) Update(ctx context.Context, id uuid.UUID, in PaymentInput, userID uuid.UUID) (*models.Payment, error) {
	if in.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("ödeme tutarı sıfır veya sıfırdan küçük olamaz")
	}

	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, err
	}

	var payment models.Payment

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		// 1. Fetch payment
		if err := txTenant.First(&payment, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPaymentNotFound
			}
			return err
		}

		if payment.Status == "canceled" {
			return errors.New("cannot_update_canceled_payment")
		}

		// 2. Revert Old Ledger Entries (revert cari, revert cash/bank, revert invoice)
		// Revert Cari Transaction
		var cariOld models.Cari
		if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&cariOld, "id = ?", payment.CariID).Error; err != nil {
			return err
		}
		var revCariType string
		var revAmountDelta decimal.Decimal
		if payment.Type == "collection" {
			revCariType = "debit"
			revAmountDelta = payment.Amount
		} else {
			revCariType = "credit"
			revAmountDelta = payment.Amount.Neg()
		}
		revCariBalanceAfter, err := UpdateCariBalance(txTenant, cariOld.ID, payment.Currency, revAmountDelta)
		if err != nil {
			return err
		}
		revCariTxID, _ := uuid.NewV7()
		revCariTx := models.CariTransaction{
			ID:           revCariTxID,
			CompanyID:    companyID,
			CariID:       payment.CariID,
			Date:         utils.NowIn(ctx),
			Type:         revCariType,
			SourceType:   "payment",
			SourceID:     &payment.ID,
			Currency:     payment.Currency,
			Amount:       payment.Amount,
			BalanceAfter: revCariBalanceAfter,
			Description:  fmt.Sprintf("Düzeltme Öncesi Geri Alma - %s (%s)", getPaymentTypeLabel(payment.Type), getMethodLabel(payment.Method)),
			CreatedBy:    &userID,
		}
		if err := txTenant.Create(&revCariTx).Error; err != nil {
			return err
		}

		// Revert Cash/Bank
		var accountBalance decimal.Decimal
		if payment.AccountKind == "cash" {
			var acc models.CashAccount
			if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&acc, "id = ?", payment.AccountID).Error; err != nil {
				return err
			}
			accountBalance = acc.Balance
		} else {
			var acc models.BankAccount
			if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&acc, "id = ?", payment.AccountID).Error; err != nil {
				return err
			}
			accountBalance = acc.Balance
		}
		var revCashType string
		var revCashBalanceAfter decimal.Decimal
		if payment.Type == "collection" {
			revCashType = "out"
			revCashBalanceAfter = accountBalance.Sub(payment.Amount)
		} else {
			revCashType = "in"
			revCashBalanceAfter = accountBalance.Add(payment.Amount)
		}
		revCashTxID, _ := uuid.NewV7()
		revCashTx := models.CashTransaction{
			ID:           revCashTxID,
			CompanyID:    companyID,
			AccountKind:  payment.AccountKind,
			AccountID:    payment.AccountID,
			Date:         utils.NowIn(ctx),
			Type:         revCashType,
			SourceType:   "payment",
			SourceID:     &payment.ID,
			Amount:       payment.Amount,
			BalanceAfter: revCashBalanceAfter,
			Description:  fmt.Sprintf("Düzeltme Öncesi Geri Alma - Cari %s: %s", getPaymentTypeLabel(payment.Type), cariOld.Name),
			CreatedBy:    &userID,
		}
		if err := txTenant.Create(&revCashTx).Error; err != nil {
			return err
		}
		if payment.AccountKind == "cash" {
			txTenant.Table("cash_accounts").Where("id = ?", payment.AccountID).Update("balance", revCashBalanceAfter)
		} else {
			txTenant.Table("bank_accounts").Where("id = ?", payment.AccountID).Update("balance", revCashBalanceAfter)
		}

		// Revert Invoice Allocation
		if payment.InvoiceID != nil {
			var lockedInvoice models.Invoice
			if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lockedInvoice, "id = ?", *payment.InvoiceID).Error; err != nil {
				return err
			}
			newPaidTotal := lockedInvoice.PaidTotal.Sub(payment.Amount)
			if newPaidTotal.LessThan(decimal.Zero) {
				newPaidTotal = decimal.Zero
			}
			lockedInvoice.PaidTotal = newPaidTotal
			var newStatus string
			if newPaidTotal.GreaterThanOrEqual(lockedInvoice.Total) {
				newStatus = "paid"
			} else if newPaidTotal.GreaterThan(decimal.Zero) {
				newStatus = "partial"
			} else {
				newStatus = "sent"
			}

			if !canTransition(lockedInvoice.Status, newStatus) {
				return fmt.Errorf("fatura statüsü '%s' iken '%s' durumuna geçişe izin verilmiyor", lockedInvoice.Status, newStatus)
			}
			lockedInvoice.Status = newStatus

			if err := txTenant.Save(&lockedInvoice).Error; err != nil {
				return err
			}
		}

		// 3. Verify and Apply NEW values
		var cariNew models.Cari
		if err := txTenant.First(&cariNew, "id = ?", in.CariID).Error; err != nil {
			return ErrCariNotFoundForPayment
		}

		var accountCurrency string
		if in.AccountKind == "cash" {
			var acc models.CashAccount
			if err := txTenant.First(&acc, "id = ?", in.AccountID).Error; err != nil {
				return ErrAccountNotFound
			}
			accountCurrency = acc.Currency
		} else {
			var acc models.BankAccount
			if err := txTenant.First(&acc, "id = ?", in.AccountID).Error; err != nil {
				return ErrAccountNotFound
			}
			accountCurrency = acc.Currency
		}

		// Ödeme dövizi hesabın dövizidir (varsayılan); base cari borcuyla tutarlı.
		currency := in.Currency
		if currency == "" {
			currency = accountCurrency
		}
		if accountCurrency != currency {
			return ErrCurrencyMismatch
		}

		if in.InvoiceID != nil {
			var invoice models.Invoice
			if err := txTenant.First(&invoice, "id = ?", *in.InvoiceID).Error; err != nil {
				return ErrInvoiceNotFoundForPay
			}
			if invoice.Currency != currency {
				return ErrCurrencyMismatch
			}
		}

		// Update payment fields
		payment.CariID = in.CariID
		payment.Type = in.Type
		payment.Date = in.Date
		payment.Method = in.Method
		payment.AccountKind = in.AccountKind
		payment.AccountID = in.AccountID
		payment.Amount = in.Amount
		payment.Currency = currency
		payment.InvoiceID = in.InvoiceID
		payment.Note = in.Note
		payment.UpdatedBy = &userID

		if in.Reference != "" {
			payment.Reference = in.Reference
		}

		if err := txTenant.Save(&payment).Error; err != nil {
			return err
		}

		// Apply NEW Ledger Entries
		var lockedCariNew models.Cari
		if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lockedCariNew, "id = ?", cariNew.ID).Error; err != nil {
			return err
		}
		var cariTxTypeNew string
		var amountDeltaNew decimal.Decimal
		if in.Type == "collection" {
			cariTxTypeNew = "credit"
			amountDeltaNew = in.Amount.Neg()
		} else {
			cariTxTypeNew = "debit"
			amountDeltaNew = in.Amount
		}
		cariBalanceAfterNew, err := UpdateCariBalance(txTenant, lockedCariNew.ID, payment.Currency, amountDeltaNew)
		if err != nil {
			return err
		}
		cariTxIDNew, _ := uuid.NewV7()
		cariTxNew := models.CariTransaction{
			ID:           cariTxIDNew,
			CompanyID:    companyID,
			CariID:       in.CariID,
			Date:         in.Date,
			Type:         cariTxTypeNew,
			SourceType:   "payment",
			SourceID:     &payment.ID,
			Currency:     payment.Currency,
			Description:  fmt.Sprintf("%s (%s) - Ref: %s (Düzenlendi)", getPaymentTypeLabel(in.Type), getMethodLabel(in.Method), payment.Reference),
			Amount:       in.Amount,
			BalanceAfter: cariBalanceAfterNew,
			CreatedBy:    &userID,
		}
		if err := txTenant.Create(&cariTxNew).Error; err != nil {
			return err
		}

		// Apply Cash/Bank
		var accountBalanceNew decimal.Decimal
		if in.AccountKind == "cash" {
			var acc models.CashAccount
			if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&acc, "id = ?", in.AccountID).Error; err != nil {
				return err
			}
			accountBalanceNew = acc.Balance
		} else {
			var acc models.BankAccount
			if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&acc, "id = ?", in.AccountID).Error; err != nil {
				return err
			}
			accountBalanceNew = acc.Balance
		}
		var cashTxTypeNew string
		var cashBalanceAfterNew decimal.Decimal
		if in.Type == "collection" {
			cashTxTypeNew = "in"
			cashBalanceAfterNew = accountBalanceNew.Add(in.Amount)
		} else {
			cashTxTypeNew = "out"
			cashBalanceAfterNew = accountBalanceNew.Sub(in.Amount)
		}
		cashTxIDNew, _ := uuid.NewV7()
		cashTxNew := models.CashTransaction{
			ID:           cashTxIDNew,
			CompanyID:    companyID,
			AccountKind:  in.AccountKind,
			AccountID:    in.AccountID,
			Date:         in.Date,
			Type:         cashTxTypeNew,
			SourceType:   "payment",
			SourceID:     &payment.ID,
			Amount:       in.Amount,
			BalanceAfter: cashBalanceAfterNew,
			Description:  fmt.Sprintf("Cari %s (Düzenlendi): %s", getPaymentTypeLabel(in.Type), lockedCariNew.Name),
			CreatedBy:    &userID,
		}
		if err := txTenant.Create(&cashTxNew).Error; err != nil {
			return err
		}
		if in.AccountKind == "cash" {
			txTenant.Table("cash_accounts").Where("id = ?", in.AccountID).Update("balance", cashBalanceAfterNew)
		} else {
			txTenant.Table("bank_accounts").Where("id = ?", in.AccountID).Update("balance", cashBalanceAfterNew)
		}

		// Apply Invoice Allocation
		if in.InvoiceID != nil {
			var lockedInvoice models.Invoice
			if err := txTenant.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lockedInvoice, "id = ?", *in.InvoiceID).Error; err != nil {
				return err
			}
			newPaidTotal := lockedInvoice.PaidTotal.Add(in.Amount)
			lockedInvoice.PaidTotal = newPaidTotal
			var newStatus string
			if newPaidTotal.GreaterThanOrEqual(lockedInvoice.Total) {
				newStatus = "paid"
			} else if newPaidTotal.GreaterThan(decimal.Zero) {
				newStatus = "partial"
			} else {
				newStatus = "sent"
			}

			if !canTransition(lockedInvoice.Status, newStatus) {
				return fmt.Errorf("fatura statüsü '%s' iken '%s' durumuna geçişe izin verilmiyor", lockedInvoice.Status, newStatus)
			}
			lockedInvoice.Status = newStatus

			if err := txTenant.Save(&lockedInvoice).Error; err != nil {
				return err
			}
		}

		summaryStr := fmt.Sprintf("%s - %s %s", getPaymentTypeLabel(payment.Type), payment.Amount.StringFixed(2), payment.Currency)
		if err := WriteAuditLog(ctx, txTenant, "payment", payment.ID, "update", userID, summaryStr); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// Helpers
func getPaymentTypeLabel(t string) string {
	if t == "collection" {
		return "Tahsilat"
	}
	return "Ödeme"
}

func getMethodLabel(m string) string {
	switch m {
	case "cash":
		return "Nakit"
	case "bank":
		return "Havale/EFT"
	case "card":
		return "Kredi Kartı"
	case "check":
		return "Çek"
	default:
		return m
	}
}

// ----------------------------------------------------
// Cash Account and Bank Account Services
// ----------------------------------------------------

type CashAccountService struct{}

func NewCashAccountService() *CashAccountService {
	return &CashAccountService{}
}

func (s *CashAccountService) Create(ctx context.Context, name string, code string, accountNo string, description string, currency string, isDefault bool, createdBy uuid.UUID) (*models.CashAccount, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found")
	}
	companyID, _ := uuid.Parse(companyIDStr.(string))

	tx := utils.GetDB(ctx, database.DB)

	// If isDefault, unset other defaults first
	if isDefault {
		tx.Model(&models.CashAccount{}).Where("company_id = ?", companyID).Update("is_default", false)
	}

	if code == "" {
		code, _ = utils.GenerateNumberWithSetting(tx, companyID, "cash_account", "transaction_prefix", "CSH")
	}

	id, _ := uuid.NewV7()
	acc := models.CashAccount{
		ID:          id,
		CompanyID:   companyID,
		Name:        name,
		Code:        code,
		AccountNo:   accountNo,
		Description: description,
		Currency:    currency,
		Balance:     decimal.Zero,
		IsDefault:   isDefault,
		CreatedBy:   &createdBy,
	}

	if err := tx.Create(&acc).Error; err != nil {
		return nil, err
	}

	if err := WriteAuditLog(ctx, tx, "cash_account", acc.ID, "create", createdBy, acc.Name); err != nil {
		return nil, err
	}

	return &acc, nil
}

func (s *CashAccountService) List(ctx context.Context) ([]models.CashAccount, error) {
	tx := utils.GetDB(ctx, database.DB)
	var list []models.CashAccount
	err := tx.Order("name ASC").Find(&list).Error
	return list, err
}

type BankAccountService struct{}

func NewBankAccountService() *BankAccountService {
	return &BankAccountService{}
}

func (s *BankAccountService) Create(ctx context.Context, name string, code string, accountNo string, description string, iban string, currency string, createdBy uuid.UUID) (*models.BankAccount, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found")
	}
	companyID, _ := uuid.Parse(companyIDStr.(string))

	tx := utils.GetDB(ctx, database.DB)

	if code == "" {
		code, _ = utils.GenerateNumberWithSetting(tx, companyID, "bank_account", "transaction_prefix", "BNK")
	}

	id, _ := uuid.NewV7()
	acc := models.BankAccount{
		ID:          id,
		CompanyID:   companyID,
		Name:        name,
		Code:        code,
		AccountNo:   accountNo,
		Description: description,
		IBAN:        iban,
		Currency:    currency,
		Balance:     decimal.Zero,
		CreatedBy:   &createdBy,
	}

	if err := tx.Create(&acc).Error; err != nil {
		return nil, err
	}

	if err := WriteAuditLog(ctx, tx, "bank_account", acc.ID, "create", createdBy, acc.Name); err != nil {
		return nil, err
	}

	return &acc, nil
}

func (s *BankAccountService) List(ctx context.Context) ([]models.BankAccount, error) {
	tx := utils.GetDB(ctx, database.DB)
	var list []models.BankAccount
	err := tx.Order("name ASC").Find(&list).Error
	return list, err
}
