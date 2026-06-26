package models

import (
	"time"

	"github.com/google/uuid"
)

type Announcement struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Title        string     `gorm:"type:varchar(255);not null" json:"title"`
	Body         string     `gorm:"type:text;not null" json:"body"`
	Category     string     `gorm:"type:varchar(32);not null;default:'bilgi'" json:"category"`
	TargetPlanID *uuid.UUID `gorm:"type:uuid" json:"target_plan_id"`
	CreatedBy    uuid.UUID  `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
}
