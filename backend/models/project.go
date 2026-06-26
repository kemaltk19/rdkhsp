package models

import (
	"github.com/google/uuid"
	"time"
)

type Project struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
	CariID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"cari_id"`
	Name         string     `gorm:"type:varchar(255);not null" json:"name"`
	Description  string     `gorm:"type:text" json:"description"`
	Code         string     `gorm:"type:varchar(50);not null;uniqueIndex:,composite:company_id" json:"code"`
	Status       string     `gorm:"type:varchar(50);not null;default:'planning'" json:"status"` // 'planning', 'in_progress', 'on_hold', 'completed', 'cancelled'
	CategoryID   *uuid.UUID `gorm:"type:uuid" json:"category_id"`
	Category     *ProjectCategory `gorm:"foreignKey:CategoryID;constraint:OnDelete:SET NULL" json:"category,omitempty"`
	StartDate    time.Time  `gorm:"not null" json:"start_date"`
	EndDate      time.Time  `gorm:"not null" json:"end_date"`
	Budget       *float64   `json:"budget"`
	Note         string     `gorm:"type:text" json:"note"`
	Invoices     []Invoice  `gorm:"many2many:project_invoices;constraint:OnDelete:CASCADE" json:"invoices,omitempty"`
	Quotes       []Quote    `gorm:"many2many:project_quotes;constraint:OnDelete:CASCADE" json:"quotes,omitempty"`
	Employees    []Employee `gorm:"many2many:project_employees;constraint:OnDelete:CASCADE" json:"employees,omitempty"`
	CreatedBy    *uuid.UUID `gorm:"type:uuid" json:"created_by"`
	UpdatedBy    *uuid.UUID `gorm:"type:uuid" json:"updated_by"`
	CreatedByUser *User    `gorm:"foreignKey:CreatedBy;constraint:OnDelete:SET NULL" json:"created_by_user,omitempty"`
	UpdatedByUser *User    `gorm:"foreignKey:UpdatedBy;constraint:OnDelete:SET NULL" json:"updated_by_user,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
