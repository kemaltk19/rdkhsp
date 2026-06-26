package services

import (
	"github.com/shopspring/decimal"

	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)

var (
	ErrExpenseNotFound          = errors.New("expense_not_found")
	ErrExpenseCategoryNotFound  = errors.New("expense_category_not_found")
	ErrExpenseAlreadyCanceled   = errors.New("expense_already_canceled")
	ErrInvalidAccountForExpense = errors.New("invalid_account_for_expense")
	ErrCariNotFoundForExpense   = errors.New("cari_not_found")
)

type ExpenseService struct{}

func NewExpenseService() *ExpenseService {
	return &ExpenseService{}
}

type ExpenseInput struct {
	CategoryID  uuid.UUID       `json:"category_id" binding:"required"`
	CariID      *uuid.UUID      `json:"cari_id"`
	Date        time.Time       `json:"date" binding:"required"`
	Description string          `json:"description" binding:"max=2000"`
	Currency    string          `json:"currency" binding:"required,max=10"`
	Amount      decimal.Decimal `json:"amount" binding:"required"`
	TaxRate     decimal.Decimal `json:"tax_rate"`     // e.g. 20 for 20%
	AccountKind *string         `json:"account_kind"` // 'cash', 'bank'
	AccountID   *uuid.UUID      `json:"account_id"`
	Status      string          `json:"status" binding:"required,oneof=paid unpaid ignored"`
	Note        string          `json:"note" binding:"max=2000"`
	IsRecurring bool            `json:"is_recurring"`
}

func (s *ExpenseService) Create(ctx context.Context, in ExpenseInput, createdBy uuid.UUID) (*models.Expense, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	companyID, _ := uuid.Parse(companyIDStr.(string))

	var expense models.Expense

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		if in.Amount.IsNegative() || in.Amount.IsZero() {
			return errors.New("gider tutarı sıfırdan büyük olmalıdır")
		}

		// 0. Verify if recurring is allowed for this category
		if in.IsRecurring {
			var count int64
			if err := txTenant.Model(&models.Expense{}).
				Where("category_id = ? AND is_recurring = ? AND status != ?", in.CategoryID, true, "canceled").
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return errors.New("bu kategoride halihazırda aktif bir tekrarlayan fiş bulunmaktadır. Aynı kategoride birden fazla tekrarlayan fiş oluşturulamaz")
			}
		}

		// 1. Verify category
		var category models.ExpenseCategory
		if err := txTenant.First(&category, "id = ?", in.CategoryID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrExpenseCategoryNotFound
			}
			return err
		}

		// 2. Verify Cari if provided
		var cari models.Cari
		if in.CariID != nil {
			if err := txTenant.First(&cari, "id = ?", *in.CariID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrCariNotFoundForExpense
				}
				return err
			}
		}

		// 3. Verify Account if paid
		if in.Status == "paid" {
			if in.AccountKind == nil || in.AccountID == nil {
				return errors.New("account_kind_and_id_required_for_paid_status")
			}
			if *in.AccountKind == "cash" {
				var acc models.CashAccount
				if err := txTenant.First(&acc, "id = ?", *in.AccountID).Error; err != nil {
					return ErrInvalidAccountForExpense
				}
			} else if *in.AccountKind == "bank" {
				var acc models.BankAccount
				if err := txTenant.First(&acc, "id = ?", *in.AccountID).Error; err != nil {
					return ErrInvalidAccountForExpense
				}
			} else {
				return errors.New("invalid_account_kind")
			}
		}

		// 4. Calculate Totals
		taxAmount := in.Amount.Mul(in.TaxRate).Div(decimal.NewFromInt(100))
		total := in.Amount.Add(taxAmount)

		expenseID, _ := uuid.NewV7()
		expense = models.Expense{
			ID:          expenseID,
			CompanyID:   companyID,
			CategoryID:  in.CategoryID,
			CariID:      in.CariID,
			Date:        in.Date,
			Description: in.Description,
			Currency:    in.Currency,
			Amount:      in.Amount,
			TaxRate:     in.TaxRate,
			TaxAmount:   taxAmount,
			Total:       total,
			AccountKind: in.AccountKind,
			AccountID:   in.AccountID,
			Status:      in.Status,
			Note:        in.Note,
			IsRecurring: in.IsRecurring,
			CreatedBy:   &createdBy,
		}

		// Save Expense
		if err := txTenant.Create(&expense).Error; err != nil {
			return err
		}

		// 5. Post Ledger Entries
		if expense.Status != "ignored" {
			if err := s.applyLedgerEntries(txTenant, &expense, &category, in.CariID, createdBy); err != nil {
				return err
			}
		}

		summaryStr := fmt.Sprintf("%s - %s %s", category.Name, expense.Amount.StringFixed(2), expense.Currency)
		if err := WriteAuditLog(ctx, txTenant, "expense", expense.ID, "create", createdBy, summaryStr); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &expense, nil
}

func (s *ExpenseService) Update(ctx context.Context, id uuid.UUID, in ExpenseInput, updatedBy uuid.UUID) (*models.Expense, error) {
	var expense models.Expense

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		if in.Amount.IsNegative() || in.Amount.IsZero() {
			return errors.New("gider tutarı sıfırdan büyük olmalıdır")
		}

		// 0. Verify if recurring is allowed for this category
		if in.IsRecurring {
			var count int64
			if err := txTenant.Model(&models.Expense{}).
				Where("category_id = ? AND is_recurring = ? AND status != ? AND id != ?", in.CategoryID, true, "canceled", id).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return errors.New("bu kategoride halihazırda aktif bir tekrarlayan fiş bulunmaktadır. Aynı kategoride birden fazla tekrarlayan fiş oluşturulamaz")
			}
		}

		// 1. Fetch current expense
		if err := txTenant.First(&expense, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrExpenseNotFound
			}
			return err
		}

		if expense.Status == "canceled" {
			return ErrExpenseAlreadyCanceled
		}

		// 2. Verify Category
		var category models.ExpenseCategory
		if err := txTenant.First(&category, "id = ?", in.CategoryID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrExpenseCategoryNotFound
			}
			return err
		}

		// 3. Verify Account if paid
		if in.Status == "paid" {
			if in.AccountKind == nil || in.AccountID == nil {
				return errors.New("account_kind_and_id_required_for_paid_status")
			}
			if *in.AccountKind == "cash" {
				var acc models.CashAccount
				if err := txTenant.First(&acc, "id = ?", *in.AccountID).Error; err != nil {
					return ErrInvalidAccountForExpense
				}
			} else if *in.AccountKind == "bank" {
				var acc models.BankAccount
				if err := txTenant.First(&acc, "id = ?", *in.AccountID).Error; err != nil {
					return ErrInvalidAccountForExpense
				}
			} else {
				return errors.New("invalid_account_kind")
			}
		}

		// 4. Revert Old Ledger Entries
		createdBy := uuid.Nil
		if expense.CreatedBy != nil {
			createdBy = *expense.CreatedBy
		}
		if err := s.revertLedgerEntries(ctx, txTenant, &expense, createdBy); err != nil {
			return err
		}

		// 5. Calculate New Totals
		taxAmount := in.Amount.Mul(in.TaxRate).Div(decimal.NewFromInt(100))
		total := in.Amount.Add(taxAmount)

		expense.CategoryID = in.CategoryID
		expense.CariID = in.CariID
		expense.Date = in.Date
		expense.Description = in.Description
		expense.Currency = in.Currency
		expense.Amount = in.Amount
		expense.TaxRate = in.TaxRate
		expense.TaxAmount = taxAmount
		expense.Total = total
		expense.AccountKind = in.AccountKind
		expense.AccountID = in.AccountID
		expense.Status = in.Status
		expense.Note = in.Note
		expense.IsRecurring = in.IsRecurring
		expense.UpdatedBy = &updatedBy

		// Save updated Expense
		if err := txTenant.Save(&expense).Error; err != nil {
			return err
		}

		// 6. Apply New Ledger Entries
		if expense.Status != "ignored" {
			if err := s.applyLedgerEntries(txTenant, &expense, &category, in.CariID, createdBy); err != nil {
				return err
			}
		}

		summaryStr := fmt.Sprintf("%s - %s %s", category.Name, expense.Amount.StringFixed(2), expense.Currency)
		if err := WriteAuditLog(ctx, txTenant, "expense", expense.ID, "update", updatedBy, summaryStr); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &expense, nil
}

func (s *ExpenseService) Cancel(ctx context.Context, id uuid.UUID, createdBy uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		// 1. Fetch current expense
		var expense models.Expense
		if err := txTenant.First(&expense, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrExpenseNotFound
			}
			return err
		}

		if expense.Status == "canceled" {
			return ErrExpenseAlreadyCanceled
		}

		// 2. Revert Ledger Entries
		if err := s.revertLedgerEntries(ctx, txTenant, &expense, createdBy); err != nil {
			return err
		}

		// 3. Update Status to canceled
		expense.Status = "canceled"
		expense.UpdatedBy = &createdBy
		if err := txTenant.Save(&expense).Error; err != nil {
			return err
		}

		summaryStr := fmt.Sprintf("%s %s", expense.Amount.StringFixed(2), expense.Currency)
		if err := WriteAuditLog(ctx, txTenant, "expense", expense.ID, "cancel", createdBy, summaryStr); err != nil {
			return err
		}

		return nil
	})
}

func (s *ExpenseService) GetByID(ctx context.Context, id uuid.UUID) (*models.Expense, error) {
	tx := utils.GetDB(ctx, database.DB)
	var expense models.Expense
	if err := tx.Preload("Category").Preload("Cari").Preload("CreatedByUser").Preload("UpdatedByUser").First(&expense, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExpenseNotFound
		}
		return nil, err
	}
	return &expense, nil
}

func (s *ExpenseService) List(ctx context.Context, page, limit int, query, sort string, filters map[string]string) ([]models.Expense, int64, error) {
	tx := utils.GetDB(ctx, database.DB)

	var list []models.Expense
	var total int64

	dbQuery := tx.Model(&models.Expense{}).Preload("Category").Preload("Cari").Preload("CreatedByUser").Preload("UpdatedByUser")

	if query != "" {
		q := "%" + query + "%"
		dbQuery = dbQuery.Where("description ILIKE ? OR note ILIKE ?", q, q)
	}

	if catFilter, exists := filters["category_id"]; exists && catFilter != "" {
		dbQuery = dbQuery.Where("category_id = ?", catFilter)
	}

	if cariFilter, exists := filters["cari_id"]; exists && cariFilter != "" {
		dbQuery = dbQuery.Where("cari_id = ?", cariFilter)
	}

	if statusFilter, exists := filters["status"]; exists && statusFilter != "" {
		dbQuery = dbQuery.Where("status = ?", statusFilter)
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
	if err := dbQuery.Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Ledger Hook Helpers

func (s *ExpenseService) applyLedgerEntries(tx *gorm.DB, expense *models.Expense, category *models.ExpenseCategory, cariID *uuid.UUID, createdBy uuid.UUID) error {
	// 1. Cash / Bank transaction (if paid)
	if expense.Status == "paid" {
		if expense.AccountKind == nil || expense.AccountID == nil {
			return errors.New("ödeme yapılan hesap bilgisi eksik")
		}
		var accountBalance decimal.Decimal
		if *expense.AccountKind == "cash" {
			var acc models.CashAccount
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&acc, "id = ?", *expense.AccountID).Error; err != nil {
				return err
			}
			accountBalance = acc.Balance
		} else {
			var acc models.BankAccount
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&acc, "id = ?", *expense.AccountID).Error; err != nil {
				return err
			}
			accountBalance = acc.Balance
		}

		newAccBalance := accountBalance.Sub(expense.Total)

		cashTxID, _ := uuid.NewV7()
		cashTx := models.CashTransaction{
			ID:           cashTxID,
			CompanyID:    expense.CompanyID,
			AccountKind:  *expense.AccountKind,
			AccountID:    *expense.AccountID,
			Date:         expense.Date,
			Type:         "out",
			SourceType:   "expense",
			SourceID:     &expense.ID,
			Amount:       expense.Total,
			BalanceAfter: newAccBalance,
			Description:  fmt.Sprintf("Gider: %s (%s)", category.Name, expense.Description),
			CreatedBy:    &createdBy,
		}

		if err := tx.Create(&cashTx).Error; err != nil {
			return err
		}

		if *expense.AccountKind == "cash" {
			if err := tx.Table("cash_accounts").Where("id = ?", *expense.AccountID).Update("balance", newAccBalance).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Table("bank_accounts").Where("id = ?", *expense.AccountID).Update("balance", newAccBalance).Error; err != nil {
				return err
			}
		}
	}

	// 2. Cari Transaction (if cari_id provided)
	if cariID != nil {
		var cari models.Cari
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&cari, "id = ?", *cariID).Error; err != nil {
			return err
		}

		if expense.Status == "paid" {
			// Write credit entry (purchase)
			balanceAfterCredit, err := UpdateCariBalance(tx, *cariID, cari.Currency, expense.Total.Neg())
			if err != nil { return err }
			creditTxID, _ := uuid.NewV7()
			creditTx := models.CariTransaction{
				ID:           creditTxID,
				CompanyID:    expense.CompanyID,
				CariID:       *cariID,
				Date:         expense.Date,
				Type:         "credit",
				SourceType:   "expense",
				Currency:     cari.Currency,
				Description:  fmt.Sprintf("Gider Fişi: %s (%s)", category.Name, expense.Description),
				Amount:       expense.Total,
				BalanceAfter: balanceAfterCredit,
				CreatedBy:    &createdBy,
			}
			if err := tx.Create(&creditTx).Error; err != nil {
				return err
			}

			// Write debit entry (immediate payment)
			balanceAfterDebit, err := UpdateCariBalance(tx, *cariID, cari.Currency, expense.Total)
			if err != nil { return err }
			debitTxID, _ := uuid.NewV7()
			debitTx := models.CariTransaction{
				ID:           debitTxID,
				CompanyID:    expense.CompanyID,
				CariID:       *cariID,
				Date:         expense.Date,
				Type:         "debit",
				SourceType:   "expense",
				Currency:     cari.Currency,
				Description:  fmt.Sprintf("Gider Ödemesi: %s (%s)", category.Name, expense.Description),
				Amount:       expense.Total,
				BalanceAfter: balanceAfterDebit,
				CreatedBy:    &createdBy,
			}
			if err := tx.Create(&debitTx).Error; err != nil {
				return err
			}

			// Cari balance already updated in UpdateCariBalance
		} else {
			// Unpaid: credit entry (purchase) only
			balanceAfterCredit, err := UpdateCariBalance(tx, *cariID, cari.Currency, expense.Total.Neg())
			if err != nil { return err }
			creditTxID, _ := uuid.NewV7()
			creditTx := models.CariTransaction{
				ID:           creditTxID,
				CompanyID:    expense.CompanyID,
				CariID:       *cariID,
				Date:         expense.Date,
				Type:         "credit",
				SourceType:   "expense",
				SourceID:     &expense.ID,
				Description:  fmt.Sprintf("Gider Fişi (Ödenmedi): %s (%s)", category.Name, expense.Description),
				Amount:       expense.Total,
				BalanceAfter: balanceAfterCredit,
				CreatedBy:    &createdBy,
			}
			if err := tx.Create(&creditTx).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *ExpenseService) revertLedgerEntries(ctx context.Context, tx *gorm.DB, expense *models.Expense, createdBy uuid.UUID) error {
	// Revert Cash/Bank if it was paid
	if expense.Status == "paid" && expense.AccountID != nil && expense.AccountKind != nil {
		var accountBalance decimal.Decimal
		if *expense.AccountKind == "cash" {
			var acc models.CashAccount
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&acc, "id = ?", *expense.AccountID).Error; err != nil {
				return err
			}
			accountBalance = acc.Balance
		} else {
			var acc models.BankAccount
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&acc, "id = ?", *expense.AccountID).Error; err != nil {
				return err
			}
			accountBalance = acc.Balance
		}

		newAccBalance := accountBalance.Add(expense.Total)

		cashTxID, _ := uuid.NewV7()
		cashTx := models.CashTransaction{
			ID:           cashTxID,
			CompanyID:    expense.CompanyID,
			AccountKind:  *expense.AccountKind,
			AccountID:    *expense.AccountID,
			Date:         utils.NowIn(ctx),
			Type:         "in",
			SourceType:   "expense",
			SourceID:     &expense.ID,
			Amount:       expense.Total,
			BalanceAfter: newAccBalance,
			Description:  fmt.Sprintf("Gider İptali/Düzeltmesi: %s", expense.Description),
			CreatedBy:    &createdBy,
		}

		if err := tx.Create(&cashTx).Error; err != nil {
			return err
		}

		if *expense.AccountKind == "cash" {
			if err := tx.Table("cash_accounts").Where("id = ?", *expense.AccountID).Update("balance", newAccBalance).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Table("bank_accounts").Where("id = ?", *expense.AccountID).Update("balance", newAccBalance).Error; err != nil {
				return err
			}
		}
	}

	// Revert Cari Transaction if cari_id was provided
	if expense.CariID != nil {
		var cari models.Cari
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&cari, "id = ?", *expense.CariID).Error; err != nil {
			return err
		}

		if expense.Status == "paid" {
			// Original had both credit (purchase) and debit (payment)
			// Reversal: Debit to reverse credit, Credit to reverse debit
			balanceAfterDebit, err := UpdateCariBalance(tx, *expense.CariID, cari.Currency, expense.Total)
			if err != nil { return err }
			debitTxID, _ := uuid.NewV7()
			debitTx := models.CariTransaction{
				ID:           debitTxID,
				CompanyID:    expense.CompanyID,
				CariID:       *expense.CariID,
				Date:         utils.NowIn(ctx),
				Type:         "debit",
				SourceType:   "expense",
				Currency:     cari.Currency,
				Description:  fmt.Sprintf("Gider Fişi İptali: %s", expense.Description),
				Amount:       expense.Total,
				BalanceAfter: balanceAfterDebit,
				CreatedBy:    &createdBy,
			}
			if err := tx.Create(&debitTx).Error; err != nil {
				return err
			}

			balanceAfterCredit, err := UpdateCariBalance(tx, *expense.CariID, cari.Currency, expense.Total.Neg())
			if err != nil { return err }
			creditTxID, _ := uuid.NewV7()
			creditTx := models.CariTransaction{
				ID:           creditTxID,
				CompanyID:    expense.CompanyID,
				CariID:       *expense.CariID,
				Date:         utils.NowIn(ctx),
				Type:         "credit",
				SourceType:   "expense",
				Currency:     cari.Currency,
				Description:  fmt.Sprintf("Gider Ödemesi İptali: %s", expense.Description),
				Amount:       expense.Total,
				BalanceAfter: balanceAfterCredit,
				CreatedBy:    &createdBy,
			}
			if err := tx.Create(&creditTx).Error; err != nil {
				return err
			}
		} else {
			// Original had credit only
			// Reversal: Debit to reverse credit, increasing Cari balance
			balanceAfterDebit, err := UpdateCariBalance(tx, *expense.CariID, cari.Currency, expense.Total)
			if err != nil { return err }
			debitTxID, _ := uuid.NewV7()
			debitTx := models.CariTransaction{
				ID:           debitTxID,
				CompanyID:    expense.CompanyID,
				CariID:       *expense.CariID,
				Date:         utils.NowIn(ctx),
				Type:         "debit",
				SourceType:   "expense",
				Currency:     cari.Currency,
				Description:  fmt.Sprintf("Gider Fişi İptali (Ödenmemiş): %s", expense.Description),
				Amount:       expense.Total,
				BalanceAfter: balanceAfterDebit,
				CreatedBy:    &createdBy,
			}
			if err := tx.Create(&debitTx).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

type RepeatExpenseItem struct {
	CategoryID   uuid.UUID       `json:"category_id" gorm:"column:category_id"`
	CategoryName string          `json:"category_name" gorm:"column:category_name"`
	Month        string          `json:"month" gorm:"column:month"`
	Count        int             `json:"count" gorm:"column:count"`
	TotalAmount  decimal.Decimal `json:"total_amount" gorm:"column:total_amount"`
	Currency     string          `json:"currency" gorm:"column:currency"`
}

func (s *ExpenseService) GetRepeatAnalysis(ctx context.Context) ([]RepeatExpenseItem, error) {
	txTenant := utils.GetDB(ctx, database.DB)

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	var results []RepeatExpenseItem
	err := txTenant.
		Table("expenses e").
		Select(`
            ec.id AS category_id,
            ec.name AS category_name,
            TO_CHAR(e.date, 'YYYY-MM') AS month,
            COUNT(e.id) AS count,
            SUM(e.total) AS total_amount,
            e.currency
        `).
		Joins("JOIN expense_categories ec ON ec.id = e.category_id").
		Where("e.date >= ? AND e.date < ? AND e.status NOT IN ('canceled', 'ignored')", startOfMonth, endOfMonth).
		Group("ec.id, ec.name, TO_CHAR(e.date, 'YYYY-MM'), e.currency").
		Having("COUNT(e.id) >= 2").
		Scan(&results).Error

	return results, err
}

// ----------------------------------------------------

// Expense Category Service
// ----------------------------------------------------

type ExpenseCategoryService struct{}

func NewExpenseCategoryService() *ExpenseCategoryService {
	return &ExpenseCategoryService{}
}

func (s *ExpenseCategoryService) CreateCategory(ctx context.Context, name string, createdBy uuid.UUID) (*models.ExpenseCategory, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found")
	}
	companyID, _ := uuid.Parse(companyIDStr.(string))

	tx := utils.GetDB(ctx, database.DB)

	id, _ := uuid.NewV7()
	cat := models.ExpenseCategory{
		ID:        id,
		CompanyID: companyID,
		Name:      name,
		CreatedBy: &createdBy,
	}

	if err := tx.Create(&cat).Error; err != nil {
		return nil, err
	}

	if err := WriteAuditLog(ctx, tx, "expense_category", cat.ID, "create", createdBy, cat.Name); err != nil {
		return nil, err
	}

	return &cat, nil
}

func (s *ExpenseCategoryService) UpdateCategory(ctx context.Context, id uuid.UUID, name string, userID uuid.UUID) (*models.ExpenseCategory, error) {
	tx := utils.GetDB(ctx, database.DB)

	var count int64
	if err := tx.Model(&models.Expense{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("category_in_use_cannot_delete")
	}

	var cat models.ExpenseCategory
	if err := tx.First(&cat, "id = ?", id).Error; err != nil {
		return nil, err
	}
	cat.Name = name
	if err := tx.Save(&cat).Error; err != nil {
		return nil, err
	}

	if err := WriteAuditLog(ctx, tx, "expense_category", cat.ID, "update", userID, cat.Name); err != nil {
		return nil, err
	}

	return &cat, nil
}

func (s *ExpenseCategoryService) DeleteCategory(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	tx := utils.GetDB(ctx, database.DB)
	var count int64
	if err := tx.Model(&models.Expense{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("category_in_use_cannot_delete")
	}

	var cat models.ExpenseCategory
	if err := tx.First(&cat, "id = ?", id).Error; err == nil {
		if err := WriteAuditLog(ctx, tx, "expense_category", cat.ID, "delete", userID, cat.Name); err != nil {
			return err
		}
	}

	return tx.Delete(&models.ExpenseCategory{}, "id = ?", id).Error
}

func (s *ExpenseCategoryService) ListCategories(ctx context.Context) ([]models.ExpenseCategory, error) {
	tx := utils.GetDB(ctx, database.DB)
	var list []models.ExpenseCategory
	err := tx.Order("name ASC").Find(&list).Error
	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		companyIDStr := ctx.Value("company_id")
		if companyIDStr != nil {
			companyID, err := uuid.Parse(companyIDStr.(string))
			if err == nil {
				var recurringCategoryIDs []uuid.UUID
				tx.Model(&models.Expense{}).
					Where("company_id = ? AND is_recurring = ? AND status != ?", companyID, true, "canceled").
					Distinct("category_id").
					Pluck("category_id", &recurringCategoryIDs)

				recurringMap := make(map[uuid.UUID]bool)
				for _, id := range recurringCategoryIDs {
					recurringMap[id] = true
				}
				for i := range list {
					if recurringMap[list[i].ID] {
						list[i].HasActiveRecurring = true
					}
				}
			}
		}
	}

	return list, nil
}
