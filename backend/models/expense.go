package models

import (
	"github.com/shopspring/decimal"

	"time"

	"github.com/google/uuid"
)

type ExpenseCategory struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
	Name      string     `gorm:"type:varchar(255);not null" json:"name"`
	CreatedBy *uuid.UUID `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	HasActiveRecurring bool `gorm:"-" json:"has_active_recurring"`
}

type Expense struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID   uuid.UUID       `gorm:"type:uuid;not null;index" json:"company_id"`
	CategoryID  uuid.UUID       `gorm:"type:uuid;not null;index" json:"category_id"`
	Category    ExpenseCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	CariID      *uuid.UUID      `gorm:"type:uuid;index" json:"cari_id"` // Nullable supplier/cari
	Cari        *Cari           `gorm:"foreignKey:CariID" json:"cari,omitempty"`
	Date        time.Time       `gorm:"not null" json:"date"`
	Description string          `gorm:"type:varchar(255)" json:"description"`
	Currency    string          `gorm:"type:varchar(10);not null;default:'TRY'" json:"currency"`
	Amount      decimal.Decimal `gorm:"type:numeric(65,4);not null;default:0" json:"amount"`
	TaxRate     decimal.Decimal `gorm:"type:numeric(5,2);not null;default:0" json:"tax_rate"` // e.g. 20.00 for 20%
	TaxAmount   decimal.Decimal `gorm:"type:numeric(65,4);not null;default:0" json:"tax_amount"`
	Total       decimal.Decimal `gorm:"type:numeric(65,4);not null;default:0" json:"total"`
	AccountKind *string         `gorm:"type:varchar(20)" json:"account_kind"`                     // 'cash', 'bank' (nullable if unpaid)
	AccountID   *uuid.UUID      `gorm:"type:uuid;index" json:"account_id"`                        // ID from cash_accounts or bank_accounts (nullable if unpaid)
	Status      string          `gorm:"type:varchar(50);not null;default:'unpaid'" json:"status"` // 'paid', 'unpaid', 'ignored', 'canceled'
	Note        string          `gorm:"type:text" json:"note"`
	IsRecurring bool            `gorm:"not null;default:false" json:"is_recurring"`
	CreatedBy   *uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	UpdatedBy   *uuid.UUID      `gorm:"type:uuid" json:"updated_by"`
	CreatedByUser *User         `gorm:"foreignKey:CreatedBy;constraint:OnDelete:SET NULL" json:"created_by_user,omitempty"`
	UpdatedByUser *User         `gorm:"foreignKey:UpdatedBy;constraint:OnDelete:SET NULL" json:"updated_by_user,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
