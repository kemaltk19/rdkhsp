package models

import (
	"github.com/shopspring/decimal"

	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID   uuid.UUID       `gorm:"type:uuid;not null;index" json:"company_id"`
	CariID      uuid.UUID       `gorm:"type:uuid;not null;index" json:"cari_id"`
	Type        string          `gorm:"type:varchar(20);not null" json:"type"` // 'collection', 'payment'
	Date        time.Time       `gorm:"not null" json:"date"`
	Method      string          `gorm:"type:varchar(50);not null" json:"method"`       // 'cash', 'bank', 'card', 'check'
	AccountKind string          `gorm:"type:varchar(20);not null" json:"account_kind"` // 'cash', 'bank'
	AccountID   uuid.UUID       `gorm:"type:uuid;not null;index" json:"account_id"`    // ID from cash_accounts or bank_accounts
	Amount      decimal.Decimal `gorm:"type:numeric(65,4);not null" json:"amount"`
	Currency    string          `gorm:"type:varchar(10);not null;default:'TRY'" json:"currency"`
	InvoiceID   *uuid.UUID      `gorm:"type:uuid;index" json:"invoice_id"` // nullable
	Reference   string          `gorm:"type:varchar(255)" json:"reference"`
	Note        string          `gorm:"type:text" json:"note"`
	Status      string          `gorm:"type:varchar(50);not null;default:'completed'" json:"status"` // 'completed', 'canceled'
	CreatedBy   *uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	UpdatedBy   *uuid.UUID      `gorm:"type:uuid" json:"updated_by"`
	CreatedByUser *User         `gorm:"foreignKey:CreatedBy;constraint:OnDelete:SET NULL" json:"created_by_user,omitempty"`
	UpdatedByUser *User         `gorm:"foreignKey:UpdatedBy;constraint:OnDelete:SET NULL" json:"updated_by_user,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
