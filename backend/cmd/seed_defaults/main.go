package main

import (
	"log"
	"os"

	"radikal-hesap/database"
	"radikal-hesap/models"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
)

func main() {
	godotenv.Load("../.env")
	database.Connect(os.Getenv("DATABASE_URL"), os.Getenv("SYSTEM_DATABASE_URL"))
	db := database.DB

	var companies []models.Company
	if err := database.SystemDB.Find(&companies).Error; err != nil {
		log.Fatalf("Failed to fetch companies: %v", err)
	}

	currenciesToSeed := []struct {
		Name   string
		Symbol string
		Code   string
	}{
		{"Türk Lirası", "₺", "TRY"},
		{"Amerikan Doları", "$", "USD"},
		{"Euro", "€", "EUR"},
		{"İngiliz Sterlini", "£", "GBP"},
		{"Rus Rublesi", "₽", "RUB"},
	}

	for _, c := range companies {
		log.Printf("Seeding defaults for company: %s (%s)", c.Name, c.ID)

		// Seed Currencies
		for i, curr := range currenciesToSeed {
			var count int64
			db.Model(&models.Currency{}).Where("company_id = ? AND code = ?", c.ID, curr.Code).Count(&count)
			if count == 0 {
				isDefault := (i == 0)
				id, _ := uuid.NewV7()
				newCurr := models.Currency{
					ID:                id,
					CompanyID:         c.ID,
					Name:              curr.Name,
					Symbol:            curr.Symbol,
					Code:              curr.Code,
					IsCrypto:          false,
					ExchangeRate:      decimal.NewFromFloat(1.0),
					FormatPosition:    "Left",
					FormatThousandSep: ",",
					FormatDecimalSep:  ".",
					FormatDecimals:    2,
					IsDefault:         isDefault,
				}
				db.Create(&newCurr)
				log.Printf("  Created currency: %s", curr.Code)
			}
		}

		// Seed Default Cash and Bank Accounts for all currencies
		for _, curr := range currenciesToSeed {
			// Cash Account
			var cashCount int64
			db.Model(&models.CashAccount{}).Where("company_id = ? AND currency = ?", c.ID, curr.Code).Count(&cashCount)
			if cashCount == 0 {
				code := "KAS-" + curr.Code
				id, _ := uuid.NewV7()
				newCash := models.CashAccount{
					ID:          id,
					CompanyID:   c.ID,
					Name:        "Merkez Kasa (" + curr.Code + ")",
					Code:        code,
					AccountNo:   "",
					Description: "Ana Nakit Kasası",
					Currency:    curr.Code,
					Balance:     decimal.Zero,
					IsDefault:   curr.Code == "TRY",
				}
				db.Create(&newCash)
				log.Printf("  Created Cash Account: %s", newCash.Name)
			}

			// Bank Account
			var bankCount int64
			db.Model(&models.BankAccount{}).Where("company_id = ? AND currency = ?", c.ID, curr.Code).Count(&bankCount)
			if bankCount == 0 {
				code := "BNK-" + curr.Code
				id, _ := uuid.NewV7()
				newBank := models.BankAccount{
					ID:          id,
					CompanyID:   c.ID,
					Name:        "Banka Hesabı (" + curr.Code + ")",
					Code:        code,
					AccountNo:   "",
					Description: "Genel Banka Hesabı",
					IBAN:        "TR000000000000000000000000",
					Currency:    curr.Code,
					Balance:     decimal.Zero,
				}
				db.Create(&newBank)
				log.Printf("  Created Bank Account: %s", newBank.Name)
			}
		}
	}

	log.Println("Seeding complete!")
}
