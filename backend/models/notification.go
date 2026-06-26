package models

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
	Type        string     `gorm:"type:varchar(50);not null" json:"type"` // invoice_dispute, quote_accepted, quote_rejected
	Title       string     `gorm:"type:varchar(255);not null" json:"title"`
	Message     string     `gorm:"type:text;not null" json:"message"`
	TargetID    *uuid.UUID `gorm:"type:uuid" json:"target_id,omitempty"`
	TargetType  string     `gorm:"type:varchar(50)" json:"target_type,omitempty"` // invoice, quote
	IsRead      bool       `gorm:"type:boolean;not null;default:false" json:"is_read"`
	CreatedAt   time.Time  `gorm:"not null" json:"created_at"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
}
