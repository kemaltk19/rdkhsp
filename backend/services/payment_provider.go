package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ChargeRequest struct {
	CompanyID  uuid.UUID
	Amount     decimal.Decimal
	Currency   string
	PlanID     uuid.UUID
	PeriodType string // "monthly" / "yearly"
}

type ChargeResult struct {
	Success     bool
	ProviderRef string
	Message     string
}

// PaymentProvider, gerçek banka/sanal POS entegrasyonunu soyutlar.
type PaymentProvider interface {
	Charge(ctx context.Context, req ChargeRequest) (*ChargeResult, error)
	VerifyWebhook(payload []byte, signature string) (bool, error)
}
