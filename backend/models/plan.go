package models

import (
	"github.com/shopspring/decimal"

	"time"

	"github.com/google/uuid"
)

type Plan struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string          `gorm:"type:varchar(255);not null" json:"name"`
	Code         string          `gorm:"type:varchar(100);not null;uniqueIndex:idx_plan_code_currency" json:"code"`
	PriceMonthly decimal.Decimal `gorm:"type:numeric(65,4);not null;default:0" json:"price_monthly"`
	PriceYearly  decimal.Decimal `gorm:"type:numeric(65,4);not null;default:0" json:"price_yearly"`
	Currency     string          `gorm:"type:varchar(10);not null;default:'TRY';uniqueIndex:idx_plan_code_currency" json:"currency"`
	Features     string          `gorm:"type:jsonb" json:"features"`
	IsActive     bool            `gorm:"not null;default:true" json:"is_active"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
