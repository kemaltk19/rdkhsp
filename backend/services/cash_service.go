package services

import (
	"fmt"
	"github.com/shopspring/decimal"

	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)

var (
	ErrCashAccountNotFound = errors.New("kasa hesabı bulunamadı")
	ErrBankAccountNotFound = errors.New("banka hesabı bulunamadı")
	ErrAccountInUse        = errors.New("hareket görmüş hesap silinemez")
	ErrTransferSameAccount = errors.New("aynı hesaba transfer yapılamaz")
)

type CashService struct{}

func NewCashService() *CashService {
	return &CashService{}
}

// ----------------------------------------------------
// Cash Account Management
// ----------------------------------------------------

func (s *CashService) UpdateCashAccount(ctx context.Context, id uuid.UUID, name string, code string, accountNo string, description string, currency string, isDefault bool, userID uuid.UUID) (*models.CashAccount, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found")
	}
	companyID, _ := uuid.Parse(companyIDStr.(string))

	tx := utils.GetDB(ctx, database.DB)

	var count int64
	if err := tx.Model(&models.CashTransaction{}).Where("company_id = ? AND account_kind = ? AND account_id = ?", companyID, "cash", id).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrAccountInUse
	}

	var acc models.CashAccount
	if err := tx.Where("id = ? AND company_id = ?", id, companyID).First(&acc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCashAccountNotFound
		}
		return nil, err
	}

	acc.Name = name
	acc.Code = code
	acc.AccountNo = accountNo
	acc.Description = description
	acc.Currency = currency
	acc.IsDefault = isDefault

	err := tx.Transaction(func(dbTx *gorm.DB) error {
		if isDefault {
			// Unset other defaults
			if err := dbTx.Model(&models.CashAccount{}).Where("company_id = ? AND id != ?", companyID, id).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		if err := dbTx.Save(&acc).Error; err != nil {
			return err
		}
		return WriteAuditLog(ctx, dbTx, "cash_account", acc.ID, "update", userID, acc.Name)
	})

	if err != nil {
		return nil, err
	}

	return &acc, nil
}

func (s *CashService) DeleteCashAccount(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return errors.New("company_id not found")
	}
	companyID, _ := uuid.Parse(companyIDStr.(string))

	tx := utils.GetDB(ctx, database.DB)

	var acc models.CashAccount
	if err := tx.Where("id = ? AND company_id = ?", id, companyID).First(&acc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCashAccountNotFound
		}
		return err
	}

	// Check if this cash account has any transactions
	var count int64
	if err := tx.Model(&models.CashTransaction{}).Where("company_id = ? AND account_kind = ? AND account_id = ?", companyID, "cash", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrAccountInUse
	}

	if err := WriteAuditLog(ctx, tx, "cash_account", acc.ID, "delete", userID, acc.Name); err != nil {
		return err
	}

	return tx.Delete(&acc).Error
}

// ----------------------------------------------------
// Bank Account Management
// ----------------------------------------------------

func (s *CashService) UpdateBankAccount(ctx context.Context, id uuid.UUID, name string, code string, accountNo string, description string, iban string, currency string, userID uuid.UUID) (*models.BankAccount, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found")
	}
	companyID, _ := uuid.Parse(companyIDStr.(string))

	tx := utils.GetDB(ctx, database.DB)

	var count int64
	if err := tx.Model(&models.CashTransaction{}).Where("company_id = ? AND account_kind = ? AND account_id = ?", companyID, "bank", id).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrAccountInUse
	}

	var acc models.BankAccount
	if err := tx.Where("id = ? AND company_id = ?", id, companyID).First(&acc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBankAccountNotFound
		}
		return nil, err
	}

	acc.Name = name
	acc.Code = code
	acc.AccountNo = accountNo
	acc.Description = description
	acc.IBAN = iban
	acc.Currency = currency

	if err := tx.Save(&acc).Error; err != nil {
		return nil, err
	}

	if err := WriteAuditLog(ctx, tx, "bank_account", acc.ID, "update", userID, acc.Name); err != nil {
		return nil, err
	}

	return &acc, nil
}

func (s *CashService) DeleteBankAccount(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return errors.New("company_id not found")
	}
	companyID, _ := uuid.Parse(companyIDStr.(string))

	tx := utils.GetDB(ctx, database.DB)

	var acc models.BankAccount
	if err := tx.Where("id = ? AND company_id = ?", id, companyID).First(&acc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBankAccountNotFound
		}
		return err
	}

	// Check if this bank account has any transactions
	var count int64
	if err := tx.Model(&models.CashTransaction{}).Where("company_id = ? AND account_kind = ? AND account_id = ?", companyID, "bank", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrAccountInUse
	}

	if err := WriteAuditLog(ctx, tx, "bank_account", acc.ID, "delete", userID, acc.Name); err != nil {
		return err
	}

	return tx.Delete(&acc).Error
}

// ----------------------------------------------------
// Money Transfer (Virman)
// ----------------------------------------------------

type TransferInput struct {
	FromKind       string          `json:"from_kind" binding:"required,oneof=cash bank"`
	FromID         uuid.UUID       `json:"from_id" binding:"required"`
	ToKind         string          `json:"to_kind" binding:"required,oneof=cash bank"`
	ToID           uuid.UUID       `json:"to_id" binding:"required"`
	Amount         decimal.Decimal `json:"amount" binding:"required"` // from_amount
	ToAmount       decimal.Decimal `json:"to_amount"`                  // optional, falls back to Amount
	ExchangeRate   decimal.Decimal `json:"exchange_rate"`             // optional, defaults to 1.0
	ExchangeRateOp string          `json:"exchange_rate_op"`          // math operator: '*', '/', '+', '-'
	Date           string          `json:"date" binding:"required"`
	Description    string          `json:"description"`
}

func (s *CashService) Transfer(ctx context.Context, in TransferInput, userID uuid.UUID) (*models.CashTransaction, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found")
	}
	companyID, _ := uuid.Parse(companyIDStr.(string))

	if in.FromID == in.ToID && in.FromKind == in.ToKind {
		return nil, ErrTransferSameAccount
	}

	if in.Amount.IsNegative() || in.Amount.IsZero() {
		return nil, errors.New("transfer tutarı sıfırdan büyük olmalıdır")
	}

	date, err := time.Parse("2006-01-02", in.Date[:10])
	if err != nil {
		date = utils.NowIn(ctx)
	}

	tx := utils.GetDB(ctx, database.DB)

	var resultTx *models.CashTransaction

	err = tx.Transaction(func(dbTx *gorm.DB) error {
		var fromCurrency, toCurrency string
		var fromName, toName string
		var fromBal, toBal decimal.Decimal

		// Load and lock source account
		if in.FromKind == "cash" {
			var acc models.CashAccount
			if err := dbTx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND company_id = ?", in.FromID, companyID).First(&acc).Error; err != nil {
				return ErrCashAccountNotFound
			}
			fromCurrency = acc.Currency
			fromName = acc.Name
			fromBal = acc.Balance
		} else {
			var acc models.BankAccount
			if err := dbTx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND company_id = ?", in.FromID, companyID).First(&acc).Error; err != nil {
				return ErrBankAccountNotFound
			}
			fromCurrency = acc.Currency
			fromName = acc.Name
			fromBal = acc.Balance
		}

		// Load and lock destination account
		if in.ToKind == "cash" {
			var acc models.CashAccount
			if err := dbTx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND company_id = ?", in.ToID, companyID).First(&acc).Error; err != nil {
				return ErrCashAccountNotFound
			}
			toCurrency = acc.Currency
			toName = acc.Name
			toBal = acc.Balance
		} else {
			var acc models.BankAccount
			if err := dbTx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND company_id = ?", in.ToID, companyID).First(&acc).Error; err != nil {
				return ErrBankAccountNotFound
			}
			toCurrency = acc.Currency
			toName = acc.Name
			toBal = acc.Balance
		}

		// Setup target amounts based on cross-currency details
		toAmount := in.Amount
		if fromCurrency != toCurrency {
			if !in.ToAmount.IsZero() {
				toAmount = in.ToAmount
			} else if !in.ExchangeRate.IsZero() {
				switch in.ExchangeRateOp {
				case "/":
					toAmount = in.Amount.Div(in.ExchangeRate)
				case "+":
					toAmount = in.Amount.Add(in.ExchangeRate)
				case "-":
					toAmount = in.Amount.Sub(in.ExchangeRate)
				default: // "*" or empty fallback
					toAmount = in.Amount.Mul(in.ExchangeRate)
				}
			} else {
				return errors.New("farklı dövizli transferlerde kur veya hedef tutar belirtilmelidir")
			}

			if toAmount.IsNegative() || toAmount.IsZero() {
				return errors.New("hedef hesaba aktarılacak tutar sıfırdan büyük olmalıdır")
			}
		}

		// Deduct from source
		fromNewBal := fromBal.Sub(in.Amount)
		if in.FromKind == "cash" {
			if err := dbTx.Model(&models.CashAccount{}).Where("id = ?", in.FromID).Update("balance", fromNewBal).Error; err != nil {
				return err
			}
		} else {
			if err := dbTx.Model(&models.BankAccount{}).Where("id = ?", in.FromID).Update("balance", fromNewBal).Error; err != nil {
				return err
			}
		}

		// Add to destination (using the calculated target amount)
		toNewBal := toBal.Add(toAmount)
		if in.ToKind == "cash" {
			if err := dbTx.Model(&models.CashAccount{}).Where("id = ?", in.ToID).Update("balance", toNewBal).Error; err != nil {
				return err
			}
		} else {
			if err := dbTx.Model(&models.BankAccount{}).Where("id = ?", in.ToID).Update("balance", toNewBal).Error; err != nil {
				return err
			}
		}

		// Generate transaction ID
		transferID, _ := uuid.NewV7()
		fromTxID, _ := uuid.NewV7()
		toTxID, _ := uuid.NewV7()

		desc := in.Description
		if desc == "" {
			desc = "Hesaplar arası transfer"
		}

		fromDesc := desc + " -> " + toName
		toDesc := desc + " <- " + fromName

		// Log source out
		fromTx := models.CashTransaction{
			ID:           fromTxID,
			CompanyID:    companyID,
			AccountKind:  in.FromKind,
			AccountID:    in.FromID,
			Date:         date,
			Type:         "out",
			SourceType:   "transfer",
			SourceID:     &transferID,
			Amount:       in.Amount,
			BalanceAfter: fromNewBal,
			Description:  fromDesc,
			CreatedBy:    &userID,
		}
		if err := dbTx.Create(&fromTx).Error; err != nil {
			return err
		}

		// Log destination in
		toTx := models.CashTransaction{
			ID:           toTxID,
			CompanyID:    companyID,
			AccountKind:  in.ToKind,
			AccountID:    in.ToID,
			Date:         date,
			Type:         "in",
			SourceType:   "transfer",
			SourceID:     &transferID,
			Amount:       toAmount,
			BalanceAfter: toNewBal,
			Description:  toDesc,
			CreatedBy:    &userID,
		}
		if err := dbTx.Create(&toTx).Error; err != nil {
			return err
		}

		resultTx = &fromTx

		summaryStr := fmt.Sprintf("%s -> %s: %s", fromName, toName, in.Amount.StringFixed(2))
		if err := WriteAuditLog(ctx, dbTx, "transfer", transferID, "create", userID, summaryStr); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return resultTx, nil
}

// ----------------------------------------------------
// Cash Transactions History Log
// ----------------------------------------------------

func (s *CashService) ListTransactions(ctx context.Context, kind string, accountID uuid.UUID) ([]models.CashTransaction, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found")
	}
	companyID, _ := uuid.Parse(companyIDStr.(string))

	tx := utils.GetDB(ctx, database.DB)

	var list []models.CashTransaction
	err := tx.Where("company_id = ? AND account_kind = ? AND account_id = ?", companyID, kind, accountID).
		Order("date DESC, created_at DESC").
		Find(&list).Error

	return list, err
}
