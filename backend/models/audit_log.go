package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
	Module    string     `gorm:"type:varchar(50);not null;index" json:"module"` // "invoice", "quote", "payment", "expense", "cari", "product", ...
	RecordID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"record_id"`
	Action    string     `gorm:"type:varchar(20);not null" json:"action"` // "create", "update", "delete", "cancel"
	UserID    *uuid.UUID `gorm:"type:uuid" json:"user_id"`
	UserName  string     `gorm:"type:varchar(255)" json:"user_name"` // snapshot — kullanıcı silinse de isim kalsın
	UserRole  string     `gorm:"type:varchar(50)" json:"user_role"`
	Summary   string     `gorm:"type:varchar(255)" json:"summary"` // örn. "FT-2026-0042 - 1.250,00 TRY"
	CreatedAt time.Time  `json:"created_at"`
}
