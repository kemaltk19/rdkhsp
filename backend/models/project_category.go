package models

import (
	"github.com/google/uuid"
	"time"
)

type ProjectCategory struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	Name      string    `gorm:"type:varchar(255);not null;uniqueIndex:,composite:company_id" json:"name"`
	Code      string    `gorm:"type:varchar(50);uniqueIndex:,composite:company_id" json:"code"`
	Color     string    `gorm:"type:varchar(10)" json:"color"`
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
