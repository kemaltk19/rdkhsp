package models

import (
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
	UserID     *uuid.UUID `gorm:"type:uuid;index" json:"user_id"` // Nullable: only if login credentials exist
	Name       string     `gorm:"type:varchar(255);not null" json:"name"`
	Email      string     `gorm:"type:varchar(255);not null" json:"email"`
	Phone      string     `gorm:"type:varchar(50)" json:"phone"`
	Position   string     `gorm:"type:varchar(255)" json:"position"`
	Department string     `gorm:"type:varchar(255)" json:"department"`
	HireDate   *time.Time `json:"hire_date"`
	IsActive   bool       `gorm:"not null;default:true" json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	User       *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
