package models

import (
	"github.com/shopspring/decimal"

	"time"

	"github.com/google/uuid"
)

type Invoice struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID     uuid.UUID       `gorm:"type:uuid;not null;index" json:"company_id"`
	CariID        uuid.UUID       `gorm:"type:uuid;not null;index" json:"cari_id"`
	Type          string          `gorm:"type:varchar(20);not null" json:"type"`    // 'sales', 'purchase'
	Number        string          `gorm:"type:varchar(100);not null" json:"number"` // unique per company sequence-generated
	Date          time.Time       `gorm:"not null" json:"date"`
	DueDate       time.Time       `gorm:"not null" json:"due_date"`
	Currency      string          `gorm:"type:varchar(10);not null;default:'TRY'" json:"currency"`
	ExchangeRate  decimal.Decimal `gorm:"type:numeric(65,10);not null;default:1" json:"exchange_rate"`
	Subtotal      decimal.Decimal `gorm:"type:numeric(65,4);not null;default:0" json:"subtotal"`
	DiscountTotal decimal.Decimal `gorm:"type:numeric(65,4);not null;default:0" json:"discount_total"`
	TaxTotal      decimal.Decimal `gorm:"type:numeric(65,4);not null;default:0" json:"tax_total"`
	Total         decimal.Decimal `gorm:"type:numeric(65,4);not null;default:0" json:"total"`
	PaidTotal     decimal.Decimal `gorm:"type:numeric(65,4);not null;default:0" json:"paid_total"`
	Status        string          `gorm:"type:varchar(50);not null;default:'draft'" json:"status"` // 'draft', 'sent', 'disputed', 'partial', 'paid', 'canceled'
	// ProcessStock false ise bu fatura stok hareketi YARATMAZ (hizmet/gider faturaları
	// gibi). Varsayılan true; iptal de bu bayrağa uyar (stok işlemeyen iptal de işlemez).
	ProcessStock  bool            `gorm:"not null;default:true" json:"process_stock"`
	Note          string          `gorm:"type:text" json:"note"`
	SentAt        *time.Time      `json:"sent_at"`
	LastSentAt    *time.Time      `json:"last_sent_at"`
	DisputeNote   string          `gorm:"type:text" json:"dispute_note"`
	DisputedAt    *time.Time      `json:"disputed_at"`
	PublicToken   string          `gorm:"type:varchar(64);uniqueIndex" json:"-"`
	Items         []InvoiceItem   `gorm:"foreignKey:InvoiceID;constraint:OnDelete:CASCADE" json:"items"`
	CreatedBy     *uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	UpdatedBy     *uuid.UUID      `gorm:"type:uuid" json:"updated_by"`
	CreatedByUser *User           `gorm:"foreignKey:CreatedBy;constraint:OnDelete:SET NULL" json:"created_by_user,omitempty"`
	UpdatedByUser *User           `gorm:"foreignKey:UpdatedBy;constraint:OnDelete:SET NULL" json:"updated_by_user,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type InvoiceItem struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID    uuid.UUID       `gorm:"type:uuid;not null;index" json:"company_id"`
	InvoiceID    uuid.UUID       `gorm:"type:uuid;not null;index" json:"invoice_id"`
	ProductID    *uuid.UUID      `gorm:"type:uuid" json:"product_id"` // nullable->products (for Module 7)
	Description  string          `gorm:"type:text" json:"description"`
	Quantity     decimal.Decimal `gorm:"type:numeric(65,4);not null;default:1" json:"quantity"`
	Unit         string          `gorm:"type:varchar(50)" json:"unit"`
	UnitPrice    decimal.Decimal `gorm:"type:numeric(65,4);not null;default:0" json:"unit_price"`
	DiscountRate decimal.Decimal `gorm:"type:numeric(5,2);not null;default:0" json:"discount_rate"`
	TaxRate      decimal.Decimal `gorm:"type:numeric(5,2);not null;default:0" json:"tax_rate"` // e.g. 20 for 20%
	LineTotal    decimal.Decimal `gorm:"type:numeric(65,4);not null;default:0" json:"line_total"` // satırın kendi dövizinde
	Currency       string          `gorm:"type:varchar(10);not null;default:'TRY'" json:"currency"`
	ExchangeRate   decimal.Decimal `gorm:"type:numeric(65,10);not null;default:1" json:"exchange_rate"`
	ExchangeRateOp string          `gorm:"type:varchar(5);not null;default:'*'" json:"exchange_rate_op"` // '*' veya '/'
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
