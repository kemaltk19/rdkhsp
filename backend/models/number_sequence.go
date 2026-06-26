package models

import "github.com/google/uuid"

type NumberSequence struct {
	CompanyID uuid.UUID `gorm:"type:uuid;primaryKey" json:"company_id"`
	Key       string    `gorm:"type:varchar(100);primaryKey" json:"key"` // e.g. 'invoice_sales', 'invoice_purchase', 'quote', 'cari'
	LastNo    int       `gorm:"not null;default:0" json:"last_no"`
}
