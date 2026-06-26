package models

import (
	"time"

	"github.com/google/uuid"
)

type Setting struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	Key       string    `gorm:"type:varchar(255);not null;index" json:"key"`
	Value     string    `gorm:"type:text" json:"value"` // Stored as string, can represent JSON or simple values
	Category  string    `gorm:"type:varchar(100);not null;default:'general'" json:"category"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
