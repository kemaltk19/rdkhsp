package main

import (
	"context"
	"fmt"
	"log"
	"radikal-hesap/config"
	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/services"

	"github.com/shopspring/decimal"
)

func main() {
	cfg := config.Load()
	database.Connect(cfg.DatabaseURL, cfg.SystemDatabaseURL)

	db := database.SystemDB // USE SYSTEM DB TO BYPASS RLS
	_ = context.Background()

	var caris []models.Cari
	if err := db.Find(&caris).Error; err != nil {
		log.Fatalf("Cariler bulunamadi: %v", err)
	}

    fmt.Printf("Bulunan cari sayisi: %d\n", len(caris))

	for _, cari := range caris {
		var txs []models.CariTransaction
		if err := db.Where("cari_id = ?", cari.ID).Find(&txs).Error; err != nil {
			continue
		}

		db.Where("cari_id = ?", cari.ID).Delete(&models.CariBalance{})

		balances := make(map[string]decimal.Decimal)
		for _, tx := range txs {
			currency := tx.Currency
			if currency == "" {
				currency = "TRY"
			}

			if tx.Type == "debit" {
				balances[currency] = balances[currency].Add(tx.Amount)
			} else {
				balances[currency] = balances[currency].Sub(tx.Amount)
			}
		}

		for currency, balance := range balances {
			if !balance.IsZero() {
				services.UpdateCariBalance(db, cari.ID, currency, balance)
			}
		}
		fmt.Printf("Cari %s (ID: %s) guncellendi\n", cari.Name, cari.ID)
	}
	fmt.Println("Bitti!")
}
