package models

import (
	"time"

	"github.com/google/uuid"
)

// Role, bir şirketin tanımladığı izin şablonudur (örn. "Satış Personeli").
// Sistem rolleri (admin/superadmin/cari) bu tabloya dahil değildir — onlar
// User.Role string alanında sabit kalır. Role tablosu sadece "personel" tipi
// kullanıcıların modüler yetkilendirmesi içindir.
type Role struct {
	ID          uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID   uuid.UUID        `gorm:"type:uuid;not null;index" json:"company_id"`
	Name        string           `gorm:"type:varchar(255);not null" json:"name"`
	Description string           `gorm:"type:text" json:"description"`
	Permissions []RolePermission `gorm:"foreignKey:RoleID" json:"permissions"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// RolePermission, bir rolün tek bir modül üzerindeki CRUD haklarını tutar.
// Bir Role'e ait modül başına tek satır vardır (role_id+module unique).
type RolePermission struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	RoleID    uuid.UUID `gorm:"type:uuid;not null;index" json:"role_id"`
	Module    string    `gorm:"type:varchar(50);not null" json:"module"` // 'caris','invoices','payments','expenses','products','reports'
	CanCreate bool      `gorm:"not null;default:false" json:"can_create"`
	CanRead   bool      `gorm:"not null;default:false" json:"can_read"`
	CanUpdate bool      `gorm:"not null;default:false" json:"can_update"`
	CanDelete bool      `gorm:"not null;default:false" json:"can_delete"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
