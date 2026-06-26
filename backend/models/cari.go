package models

import (
	"time"

	"github.com/google/uuid"
)

type Cari struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	Code      string    `gorm:"type:varchar(100);not null" json:"code"`
	Type      string    `gorm:"type:varchar(50);not null" json:"type"` // 'customer', 'supplier', 'both', or custom
	Group     string    `gorm:"type:varchar(100)" json:"group"`
	Title     string    `gorm:"type:varchar(255)" json:"title"`        // Resmi ünvan (opsiyonel)
	Name      string    `gorm:"type:varchar(255)" json:"name"`         // Firma adı / cari adı
	ContactName string  `gorm:"type:varchar(255)" json:"contact_name"` // Ad Soyad / yetkili kişi
	TaxOffice string    `gorm:"type:varchar(255)" json:"tax_office"`
	TaxNumber string    `gorm:"type:varchar(50)" json:"tax_number"`
	Email     string    `gorm:"type:varchar(255)" json:"email"`
	Phone     string    `gorm:"type:varchar(50)" json:"phone"`
	Landline  string    `gorm:"type:varchar(50)" json:"landline"`
	Fax       string    `gorm:"type:varchar(50)" json:"fax"`

	// Fatura adresi
	Address    string `gorm:"type:text" json:"address"`
	City       string `gorm:"type:varchar(100)" json:"city"`     // İl
	District   string `gorm:"type:varchar(100)" json:"district"` // İlçe
	PostalCode string `gorm:"type:varchar(20)" json:"postal_code"`
	Country    string `gorm:"type:varchar(100);not null;default:'Türkiye'" json:"country"`

	// Sevk adresi
	ShippingAddress    string `gorm:"type:text" json:"shipping_address"`
	ShippingCity       string `gorm:"type:varchar(100)" json:"shipping_city"`
	ShippingDistrict   string `gorm:"type:varchar(100)" json:"shipping_district"`
	ShippingPostalCode string `gorm:"type:varchar(20)" json:"shipping_postal_code"`
	ShippingCountry    string `gorm:"type:varchar(100)" json:"shipping_country"`

	Currency  string        `gorm:"type:varchar(10);not null;default:'TRY'" json:"currency"`
	Balances  []CariBalance `gorm:"foreignKey:CariID" json:"balances"`
	Persons   []CariPerson  `gorm:"foreignKey:CariID" json:"persons"`
	IsActive  bool          `gorm:"not null;default:true" json:"is_active"`
	Note      string        `gorm:"type:text" json:"note"`
	CreatedBy *uuid.UUID    `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type CariPerson struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CariID    uuid.UUID `gorm:"type:uuid;not null;index" json:"cari_id"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	Title     string    `gorm:"type:varchar(100)" json:"title"`   // Unvan (Müdür, Muhasebe vb.)
	Phone     string    `gorm:"type:varchar(50)" json:"phone"`    // Kişi cep telefonu
	Email     string    `gorm:"type:varchar(255)" json:"email"`   // Kişi e-postası
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (CariPerson) TableName() string {
	return "cari_persons"
}
