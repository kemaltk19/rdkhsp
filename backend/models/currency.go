package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Currency struct {
	ID                 uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID          uuid.UUID       `gorm:"type:uuid;not null;index" json:"company_id"`
	Name               string          `gorm:"type:varchar(100);not null" json:"name"`
	Symbol             string          `gorm:"type:varchar(10);not null" json:"symbol"`
	Code               string          `gorm:"type:varchar(10);not null" json:"code"`
	IsCrypto           bool            `gorm:"default:false" json:"is_crypto"`
	ExchangeRate       decimal.Decimal `gorm:"type:numeric(65,4);default:1.0" json:"exchange_rate"`
	ExchangeRateOp     string          `gorm:"type:varchar(5);default:'*'" json:"exchange_rate_op"` // '*' veya '/'
	FormatPosition     string          `gorm:"type:varchar(20);default:'Left'" json:"format_position"` // Left, Right, LeftSpace, RightSpace
	FormatThousandSep  string          `gorm:"type:varchar(5);default:','" json:"format_thousand_sep"`
	FormatDecimalSep   string          `gorm:"type:varchar(5);default:'.'" json:"format_decimal_sep"`
	FormatDecimals     int             `gorm:"default:2" json:"format_decimals"`
	IsDefault          bool            `gorm:"default:false" json:"is_default"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
