package models

import (
	"github.com/shopspring/decimal"
	"time"

	"github.com/google/uuid"
)

type CariBalance struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	CariID    uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_cari_currency" json:"cari_id"`
	Currency  string          `gorm:"type:varchar(10);not null;uniqueIndex:idx_cari_currency" json:"currency"`
	Balance   decimal.Decimal `gorm:"type:numeric(65,4);not null;default:0" json:"balance"`
	UpdatedAt time.Time       `json:"updated_at"`
}
