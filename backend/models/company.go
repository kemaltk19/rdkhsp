package models

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name                 string     `gorm:"type:varchar(255);not null" json:"name"`
	ContactName          string     `gorm:"type:varchar(255)" json:"contact_name"`
	Slug                 string     `gorm:"type:varchar(255);not null" json:"slug"`
	Email                string     `gorm:"type:varchar(255)" json:"email"`
	Phone                string     `gorm:"type:varchar(50)" json:"phone"`
	Landline             string     `gorm:"type:varchar(50)" json:"landline"`
	Fax                  string     `gorm:"type:varchar(50)" json:"fax"`
	TaxOffice            string     `gorm:"type:varchar(255)" json:"tax_office"`
	TaxNumber            string     `gorm:"type:varchar(50)" json:"tax_number"`
	Industry             string     `gorm:"type:varchar(255)" json:"industry"`
	Country              string     `gorm:"type:varchar(255)" json:"country"`
	City                 string     `gorm:"type:varchar(255)" json:"city"`
	District             string     `gorm:"type:varchar(255)" json:"district"`
	Address              string     `gorm:"type:text" json:"address"`
	LogoURL              string     `gorm:"type:text" json:"logo_url"`
	Currency             string     `gorm:"type:varchar(10);not null;default:'TRY'" json:"currency"`
	Locale               string     `gorm:"type:varchar(10);not null;default:'tr'" json:"locale"`
	Timezone             string     `gorm:"type:varchar(64);not null;default:'Europe/Istanbul'" json:"timezone"`
	SubscriptionStatus   string     `gorm:"type:varchar(50);not null;default:'trial'" json:"subscription_status"` // 'trial', 'active', 'past_due', 'canceled'
	PlanID               *uuid.UUID `gorm:"type:uuid" json:"plan_id"`
	TrialEndsAt          time.Time  `json:"trial_ends_at"`
	CurrentPeriodEnd     *time.Time `json:"current_period_end"`
	StripeCustomerID     string     `gorm:"type:varchar(255)" json:"stripe_customer_id"`
	StripeSubscriptionID string     `gorm:"type:varchar(255)" json:"stripe_subscription_id"`
	EnabledModules       *string    `gorm:"type:jsonb" json:"enabled_modules"`
	ProjectCodePrefix    string     `gorm:"type:varchar(10);not null;default:'prj-'" json:"project_code_prefix"`
	ProjectCodeCounter   int64      `gorm:"not null;default:-1" json:"project_code_counter"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}
