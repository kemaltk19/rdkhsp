package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)

var (
	ErrCurrencyNotFound     = errors.New("currency_not_found")
	ErrCurrencyDefaultDelete = errors.New("default_currency_cannot_be_deleted")
)

type CurrencyService struct{}

func NewCurrencyService() *CurrencyService {
	return &CurrencyService{}
}

type CurrencyInput struct {
	Name              string          `json:"name" binding:"required"`
	Code              string          `json:"code" binding:"required"`
	Symbol            string          `json:"symbol"`
	IsCrypto          bool            `json:"is_crypto"`
	ExchangeRate      decimal.Decimal `json:"exchange_rate"`
	ExchangeRateOp    string          `json:"exchange_rate_op"`
	FormatPosition    string          `json:"format_position"`
	FormatThousandSep string          `json:"format_thousand_sep"`
	FormatDecimalSep  string          `json:"format_decimal_sep"`
	FormatDecimals    int             `json:"format_decimals"`
	IsDefault         bool            `json:"is_default"`
}

func companyIDFromCtx(ctx context.Context) (uuid.UUID, error) {
	v := ctx.Value("company_id")
	if v == nil {
		return uuid.Nil, errors.New("company_id not found in context")
	}
	return uuid.Parse(v.(string))
}

func (s *CurrencyService) List(ctx context.Context) ([]models.Currency, error) {
	tx := utils.GetDB(ctx, database.DB)
	var currencies []models.Currency
	if err := tx.Order("is_default desc, code asc").Find(&currencies).Error; err != nil {
		return nil, err
	}
	return currencies, nil
}

func (s *CurrencyService) Create(ctx context.Context, in CurrencyInput) (*models.Currency, error) {
	companyID, err := companyIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	var currency models.Currency
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		// If marked default, clear other defaults first
		if in.IsDefault {
			if err := txTenant.Model(&models.Currency{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
				return err
			}
		}

		exchangeRate := in.ExchangeRate
		if exchangeRate.IsZero() {
			exchangeRate = decimal.NewFromInt(1)
		}
		// Default (base) para biriminin base'e kuru tanımı gereği 1.0'dır.
		if in.IsDefault {
			exchangeRate = decimal.NewFromInt(1)
		}
		currency = models.Currency{
			ID:                uuid.New(),
			CompanyID:         companyID,
			Name:              in.Name,
			Code:              in.Code,
			Symbol:            in.Symbol,
			IsCrypto:          in.IsCrypto,
			ExchangeRate:      exchangeRate,
			ExchangeRateOp:    defaultStr(in.ExchangeRateOp, "*"),
			FormatPosition:    defaultStr(in.FormatPosition, "RightSpace"),
			FormatThousandSep: defaultStr(in.FormatThousandSep, "."),
			FormatDecimalSep:  defaultStr(in.FormatDecimalSep, ","),
			FormatDecimals:    in.FormatDecimals,
			IsDefault:         in.IsDefault,
		}
		return txTenant.Create(&currency).Error
	})
	if err != nil {
		return nil, err
	}
	// Default seçildiyse base tek kaynağı (Company.Currency) bununla eşitle.
	if in.IsDefault {
		if cid, e := companyIDFromCtx(ctx); e == nil {
			database.SystemDB.Model(&models.Company{}).Where("id = ?", cid).Update("currency", currency.Code)
		}
	}
	return &currency, nil
}

func (s *CurrencyService) Update(ctx context.Context, id uuid.UUID, in CurrencyInput) (*models.Currency, error) {
	var currency models.Currency
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		if err := txTenant.First(&currency, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCurrencyNotFound
			}
			return err
		}

		if in.IsDefault {
			if err := txTenant.Model(&models.Currency{}).Where("id <> ?", id).Update("is_default", false).Error; err != nil {
				return err
			}
		}

		currency.Name = in.Name
		currency.Code = in.Code
		currency.Symbol = in.Symbol
		currency.IsCrypto = in.IsCrypto
		if !in.ExchangeRate.IsZero() {
			currency.ExchangeRate = in.ExchangeRate
		}
		// Default (base) para biriminin base'e kuru tanımı gereği 1.0'dır.
		if in.IsDefault {
			currency.ExchangeRate = decimal.NewFromInt(1)
		}
		currency.ExchangeRateOp = defaultStr(in.ExchangeRateOp, currency.ExchangeRateOp)
		currency.FormatPosition = defaultStr(in.FormatPosition, currency.FormatPosition)
		currency.FormatThousandSep = defaultStr(in.FormatThousandSep, currency.FormatThousandSep)
		currency.FormatDecimalSep = defaultStr(in.FormatDecimalSep, currency.FormatDecimalSep)
		currency.FormatDecimals = in.FormatDecimals
		currency.IsDefault = in.IsDefault

		return txTenant.Save(&currency).Error
	})
	if err != nil {
		return nil, err
	}
	// Default seçildiyse base tek kaynağı (Company.Currency) bununla eşitle.
	if in.IsDefault {
		if cid, e := companyIDFromCtx(ctx); e == nil {
			database.SystemDB.Model(&models.Company{}).Where("id = ?", cid).Update("currency", currency.Code)
		}
	}
	return &currency, nil
}

func (s *CurrencyService) Delete(ctx context.Context, id uuid.UUID) error {
	tx := utils.GetDB(ctx, database.DB)

	var currency models.Currency
	if err := tx.First(&currency, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCurrencyNotFound
		}
		return err
	}
	if currency.IsDefault {
		return ErrCurrencyDefaultDelete
	}
	return tx.Delete(&currency).Error
}

func defaultStr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

// GetCurrencyRateToBase, verilen para biriminin admin Para Birimi tablosundaki
// base'e (şirket varsayılan dövizi) çevrim kurunu ve işlemini döndürür.
// Base ile aynıysa, kayıt yoksa veya kur 0 ise güvenli varsayılan (1, "*") döner.
// tx tenant-scoped olmalıdır (Currency tablosu RLS ile şirkete bağlıdır).
func GetCurrencyRateToBase(tx *gorm.DB, baseCurrency, code string) (decimal.Decimal, string) {
	if code == "" || code == baseCurrency {
		return decimal.NewFromInt(1), "*"
	}
	var c models.Currency
	if err := tx.Where("code = ?", code).First(&c).Error; err != nil {
		return decimal.NewFromInt(1), "*"
	}
	if c.ExchangeRate.IsZero() {
		return decimal.NewFromInt(1), "*"
	}
	op := c.ExchangeRateOp
	if op == "" {
		op = "*"
	}
	return c.ExchangeRate, op
}

// SyncDefaultCurrency, Company.Currency (tek base kaynağı) ile Currency tablosundaki
// IsDefault bayrağını eşitler ve base para biriminin kurunu 1.0/"*"e sabitler.
// Base koda ait Currency satırı yoksa dokunmaz (kullanıcı sonradan ekleyebilir).
func SyncDefaultCurrency(tx *gorm.DB, baseCurrency string) error {
	if baseCurrency == "" {
		return nil
	}
	if err := tx.Model(&models.Currency{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
		return err
	}
	var c models.Currency
	if err := tx.Where("code = ?", baseCurrency).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	return tx.Model(&c).Updates(map[string]interface{}{
		"is_default":       true,
		"exchange_rate":    decimal.NewFromInt(1),
		"exchange_rate_op": "*",
	}).Error
}
