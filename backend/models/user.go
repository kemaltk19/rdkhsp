package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID    *uuid.UUID `gorm:"type:uuid" json:"company_id"`
	Name         string     `gorm:"type:varchar(255);not null" json:"name"`
	Email        string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	Role         string     `gorm:"type:varchar(50);not null" json:"role"` // 'superadmin', 'admin', 'personel', 'cari'
	RoleID       *uuid.UUID `gorm:"type:uuid;index" json:"role_id"`        // sadece Role='personel' iken anlamlı; modül izin şablonu
	RoleRef      *Role      `gorm:"foreignKey:RoleID" json:"role_ref,omitempty"`
	CariID       *uuid.UUID `gorm:"type:uuid" json:"cari_id"`
	Locale       string     `gorm:"type:varchar(10);not null;default:'tr'" json:"locale"`
	Timezone     string     `gorm:"type:varchar(64);not null;default:'Europe/Istanbul'" json:"timezone"`
	Phone        string     `gorm:"type:varchar(50)" json:"phone"`
	IsActive     bool       `gorm:"not null;default:true" json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
