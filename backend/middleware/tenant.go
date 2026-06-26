package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"radikal-hesap/database"
	"radikal-hesap/utils"
)

// TenantMiddleware opens a database transaction and configures the PostgreSQL RLS context
// parameter `app.company_id` for the duration of the request.
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		companyIDStr, exists := c.Get("company_id")
		if !exists || companyIDStr == "" {
			companyIDStr = "00000000-0000-0000-0000-000000000000"
		}

		// Open transaction on the main (RLS) database
		tx := database.DB.Begin()
		if tx.Error != nil {
			utils.Err(c, http.StatusInternalServerError, "INTERNAL", "Veritabanı hatası", nil)
			c.Abort()
			return
		}

		// Set the tenant company_id for this transaction (parametreli; SQL enjeksiyonuna kapali)
		if err := tx.Exec("SELECT set_config('app.company_id', ?, true)", companyIDStr).Error; err != nil {
			tx.Rollback()
			utils.Err(c, http.StatusInternalServerError, "INTERNAL", "Tenant baglanti hatasi", nil)
			c.Abort()
			return
		}

		// Retrieve the company's timezone
		companyTimezone := "Europe/Istanbul"
		if companyIDStr != "00000000-0000-0000-0000-000000000000" {
			_ = tx.Table("companies").Select("timezone").Where("id = ?", companyIDStr).Row().Scan(&companyTimezone)
		}
		if companyTimezone == "" {
			companyTimezone = "Europe/Istanbul"
		}
		loc := utils.LoadLocation(companyTimezone)

		// Put the transaction in the Gin Context and Go Context
		c.Set("db_tx", tx)
		ctx := context.WithValue(c.Request.Context(), utils.TxKey, tx)
		ctx = context.WithValue(ctx, utils.LocationKey, loc)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// Commit or Rollback based on status/errors
		if len(c.Errors) > 0 || c.Writer.Status() >= 400 {
			tx.Rollback()
		} else {
			if err := tx.Commit().Error; err != nil {
				tx.Rollback()
			}
		}
	}
}
