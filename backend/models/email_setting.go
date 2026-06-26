package models

import (
	"time"

	"github.com/google/uuid"
)

// EmailSetting holds the single, platform-wide SMTP configuration managed by the
// superadmin. All tenants/admins use this one SMTP. Global (no company_id / RLS);
// managed via SystemDB. Password stored AES-GCM encrypted.
type EmailSetting struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Host        string    `gorm:"type:varchar(255)" json:"host"`
	Port        string    `gorm:"type:varchar(10)" json:"port"`
	Username    string    `gorm:"type:varchar(255)" json:"username"`
	PasswordEnc string    `gorm:"type:text" json:"-"` // AES-GCM encrypted, never serialized
	FromEmail   string    `gorm:"type:varchar(255)" json:"from_email"`
	FromName    string    `gorm:"type:varchar(255)" json:"from_name"`
	Enabled     bool      `gorm:"default:false" json:"enabled"`
	UpdatedAt   time.Time `json:"updated_at"`
}
