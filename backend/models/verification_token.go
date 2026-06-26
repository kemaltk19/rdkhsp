package models

import (
	"time"

	"github.com/google/uuid"
)

// VerificationToken backs password reset and email verification flows.
// No company_id -> no RLS (managed via SystemDB, like RefreshToken).
type VerificationToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Type      string     `gorm:"type:varchar(30);not null;index" json:"type"` // password_reset | email_verify
	TokenHash string     `gorm:"type:varchar(64);not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}
