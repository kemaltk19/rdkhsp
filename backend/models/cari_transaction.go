package models

import (
	"github.com/shopspring/decimal"

	"time"

	"github.com/google/uuid"
)

type CariTransaction struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID    uuid.UUID       `gorm:"type:uuid;not null;index" json:"company_id"`
	CariID       uuid.UUID       `gorm:"type:uuid;not null;index" json:"cari_id"`
	Date         time.Time       `gorm:"not null" json:"date"`
	Type         string          `gorm:"type:varchar(20);not null" json:"type"`        // 'debit' (borc), 'credit' (alacak)
	SourceType   string          `gorm:"type:varchar(50);not null" json:"source_type"` // 'invoice', 'payment', 'expense', 'manual'
	SourceID     *uuid.UUID      `gorm:"type:uuid" json:"source_id"`
	Description  string          `gorm:"type:varchar(255)" json:"description"`
	Currency     string          `gorm:"type:varchar(10);not null;default:'TRY'" json:"currency"`
	Amount       decimal.Decimal `gorm:"type:numeric(65,4);not null" json:"amount"`
	BalanceAfter decimal.Decimal `gorm:"type:numeric(65,4);not null" json:"balance_after"`
	CreatedBy    *uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
