package models

import (
	"github.com/shopspring/decimal"

	"time"

	"github.com/google/uuid"
)

type ProductCategory struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID      uuid.UUID       `gorm:"type:uuid;not null;index" json:"company_id"`
	Name           string          `gorm:"type:varchar(255);not null" json:"name"`
	DefaultKDVRate decimal.Decimal `gorm:"type:numeric(5,2);not null;default:20" json:"default_kdv_rate"`
	CreatedBy      *uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Product struct {
	ID            uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID     uuid.UUID        `gorm:"type:uuid;not null;index" json:"company_id"`
	Code          string           `gorm:"type:varchar(100);not null" json:"code"` // unique per company sequence-generated
	Name          string           `gorm:"type:varchar(255);not null" json:"name"`
	Brand         string           `gorm:"type:varchar(255)" json:"brand"`
	Type          string           `gorm:"type:varchar(50);not null;default:'product'" json:"type"` // 'product', 'service'
	Unit          string           `gorm:"type:varchar(50)" json:"unit"`                            // Adet, Kg, etc.
	Barcode       string           `gorm:"type:varchar(100)" json:"barcode"`
	CustomCodes   string           `gorm:"type:varchar(500)" json:"custom_codes"`
	SerialNumbers string           `gorm:"type:text" json:"serial_numbers"`
	PurchasePrice decimal.Decimal  `gorm:"type:numeric(65,4);not null;default:0" json:"purchase_price"`
	AverageCost   decimal.Decimal  `gorm:"type:numeric(65,4);not null;default:0" json:"average_cost"`
	SalePrice     decimal.Decimal  `gorm:"type:numeric(65,4);not null;default:0" json:"sale_price"`
	Currency      string           `gorm:"type:varchar(10);not null;default:'TRY'" json:"currency"`
	TaxIncluded   bool             `gorm:"not null;default:false" json:"tax_included"`          // satış fiyatı KDV dahil mi
	PurchaseTaxIncluded bool       `gorm:"not null;default:false" json:"purchase_tax_included"` // alış fiyatı KDV dahil mi
	TaxRate       decimal.Decimal  `gorm:"type:numeric(5,2);not null;default:20" json:"tax_rate"`          // satış KDV oranı, e.g. 20 for 20%
	PurchaseTaxRate decimal.Decimal `gorm:"type:numeric(5,2);not null;default:20" json:"purchase_tax_rate"` // alış KDV oranı
	Description   string           `gorm:"type:varchar(500)" json:"description"`
	TrackStock    bool             `gorm:"not null;default:true" json:"track_stock"`
	CurrentStock  decimal.Decimal  `gorm:"type:numeric(65,4);not null;default:0" json:"current_stock"` // cached stock quantity
	MinStock      decimal.Decimal  `gorm:"type:numeric(65,4);not null;default:5" json:"min_stock"`
	CategoryID    *uuid.UUID       `gorm:"type:uuid;index" json:"category_id"`
	Category      *ProductCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	IsActive      bool             `gorm:"not null;default:true" json:"is_active"`
	CreatedBy     *uuid.UUID       `gorm:"type:uuid" json:"created_by"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

type Warehouse struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
	Name      string     `gorm:"type:varchar(255);not null" json:"name"`
	Address   string     `gorm:"type:text" json:"address"`
	IsDefault bool       `gorm:"not null;default:false" json:"is_default"`
	CreatedBy *uuid.UUID `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type StockMovement struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID    uuid.UUID       `gorm:"type:uuid;not null;index" json:"company_id"`
	ProductID    uuid.UUID       `gorm:"type:uuid;not null;index" json:"product_id"`
	Product      *Product        `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	WarehouseID  uuid.UUID       `gorm:"type:uuid;not null;index" json:"warehouse_id"`
	Warehouse    *Warehouse      `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Date         time.Time       `gorm:"not null" json:"date"`
	Type         string          `gorm:"type:varchar(20);not null" json:"type"`        // 'in', 'out', 'transfer', 'adjustment'
	SourceType   string          `gorm:"type:varchar(50);not null" json:"source_type"` // 'invoice', 'manual', 'transfer'
	SourceID     *uuid.UUID      `gorm:"type:uuid" json:"source_id"`
	Quantity     decimal.Decimal `gorm:"type:numeric(65,4);not null" json:"quantity"`
	UnitCost     decimal.Decimal `gorm:"type:numeric(65,4);not null;default:0" json:"unit_cost"`
	BalanceAfter decimal.Decimal `gorm:"type:numeric(65,4);not null" json:"balance_after"`
	Note         string          `gorm:"type:varchar(255)" json:"note"`
	CreatedBy    *uuid.UUID      `gorm:"type:uuid" json:"created_by"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
