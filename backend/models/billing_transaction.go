package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BillingTransaction struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID   uuid.UUID       `gorm:"type:uuid;not null;index" json:"company_id"`
	PlanID      *uuid.UUID      `gorm:"type:uuid" json:"plan_id"`
	Action      string          `gorm:"type:varchar(20);not null" json:"action"` // "subscribe", "renew", "cancel"
	Amount      decimal.Decimal `gorm:"type:numeric(65,4);not null;default:0" json:"amount"`
	Currency    string          `gorm:"type:varchar(10);not null" json:"currency"`
	Status      string          `gorm:"type:varchar(20);not null" json:"status"` // "success", "failed", "pending"
	ProviderRef string          `gorm:"type:varchar(255)" json:"provider_ref"`  // banka API işlem referansı
	EventID     string          `gorm:"type:varchar(255);uniqueIndex" json:"event_id"` // webhook event id — idempotency (çift işlemeyi engeller)
	PeriodStart time.Time       `json:"period_start"`
	PeriodEnd   time.Time       `json:"period_end"`
	CreatedAt   time.Time       `json:"created_at"`
}
