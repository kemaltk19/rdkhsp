package models

import (
	"github.com/shopspring/decimal"

	"time"

	"github.com/google/uuid"
)

type CashAccount struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID   uuid.UUID       `gorm:"type:uuid;not null;index" json:"company_id"`
	Code        string          `gorm:"type:varchar(50)" json:"code"`
	Name        string          `gorm:"type:varchar(255);not null" json:"name"`
	AccountNo   string          `gorm:"type:varchar(100)" json:"account_no"`
	Description string          `gorm:"type:varchar(255)" json:"description"`
	Currency    string          `gorm:"type:varchar(10);not null;default:'TRY'" json:"currency"`
	Balance     decimal.Decimal `gorm:"type:numeric(65,4);not null;default:0" json:"balance"`
	IsDefault bool            `gorm:"not null;default:false" json:"is_default"`
	CreatedBy *uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type BankAccount struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID   uuid.UUID       `gorm:"type:uuid;not null;index" json:"company_id"`
	Code        string          `gorm:"type:varchar(50)" json:"code"`
	Name        string          `gorm:"type:varchar(255);not null" json:"name"`
	AccountNo   string          `gorm:"type:varchar(100)" json:"account_no"`
	IBAN        string          `gorm:"type:varchar(100)" json:"iban"`
	Description string          `gorm:"type:varchar(255)" json:"description"`
	Currency    string          `gorm:"type:varchar(10);not null;default:'TRY'" json:"currency"`
	Balance     decimal.Decimal `gorm:"type:numeric(65,4);not null;default:0" json:"balance"`
	CreatedBy   *uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type CashTransaction struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID    uuid.UUID       `gorm:"type:uuid;not null;index" json:"company_id"`
	AccountKind  string          `gorm:"type:varchar(20);not null" json:"account_kind"` // 'cash', 'bank'
	AccountID    uuid.UUID       `gorm:"type:uuid;not null;index" json:"account_id"`
	Date         time.Time       `gorm:"not null" json:"date"`
	Type         string          `gorm:"type:varchar(20);not null" json:"type"`        // 'in', 'out'
	SourceType   string          `gorm:"type:varchar(50);not null" json:"source_type"` // 'payment', 'expense', 'manual', 'transfer'
	SourceID     *uuid.UUID      `gorm:"type:uuid" json:"source_id"`
	Amount       decimal.Decimal `gorm:"type:numeric(65,4);not null" json:"amount"`
	BalanceAfter decimal.Decimal `gorm:"type:numeric(65,4);not null" json:"balance_after"`
	Description  string          `gorm:"type:varchar(255)" json:"description"`
	CreatedBy    *uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
