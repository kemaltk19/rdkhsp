package models

import (
	"time"

	"github.com/google/uuid"
)

// AnnouncementCategory, superadmin tarafından yönetilen platform-geneli duyuru
// kategorisidir. Slug (örn. "bilgi") Announcement.Category alanıyla eşleşir.
type AnnouncementCategory struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Slug      string    `gorm:"type:varchar(32);not null;uniqueIndex" json:"slug"`
	Name      string    `gorm:"type:varchar(64);not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
