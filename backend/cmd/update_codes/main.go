package main

import (
	"fmt"
	"log"
	"os"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	dsn := os.Getenv("DATABASE_URL")
	sysDsn := os.Getenv("SYSTEM_DATABASE_URL")
	database.Connect(dsn, sysDsn)
	db := database.SystemDB

	// Get all cash accounts with empty code
	var cashAccounts []models.CashAccount
	if err := db.Where("code = '' OR code IS NULL").Find(&cashAccounts).Error; err != nil {
		log.Fatal(err)
	}

	for _, acc := range cashAccounts {
		code, err := utils.GenerateNumber(db, acc.CompanyID, "cash_account", "CSH")
		if err != nil {
			log.Printf("Failed to generate code for cash account %s: %v\n", acc.ID, err)
			continue
		}
		if err := db.Model(&acc).Update("code", code).Error; err != nil {
			log.Printf("Failed to update cash account %s: %v\n", acc.ID, err)
		} else {
			fmt.Printf("Updated Cash Account %s with code %s\n", acc.Name, code)
		}
	}

	// Get all bank accounts with empty code
	var bankAccounts []models.BankAccount
	if err := db.Where("code = '' OR code IS NULL").Find(&bankAccounts).Error; err != nil {
		log.Fatal(err)
	}

	for _, acc := range bankAccounts {
		code, err := utils.GenerateNumber(db, acc.CompanyID, "bank_account", "BANK")
		if err != nil {
			log.Printf("Failed to generate code for bank account %s: %v\n", acc.ID, err)
			continue
		}
		if err := db.Model(&acc).Update("code", code).Error; err != nil {
			log.Printf("Failed to update bank account %s: %v\n", acc.ID, err)
		} else {
			fmt.Printf("Updated Bank Account %s with code %s\n", acc.Name, code)
		}
	}

	fmt.Println("Done updating codes.")
}
