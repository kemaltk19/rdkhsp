package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)

// SubscriptionMiddleware verifies that the tenant has a valid trial or active subscription.
func SubscriptionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		companyIDStr, exists := c.Get("company_id")
		if !exists || companyIDStr == "" {
			companyIDStr = "00000000-0000-0000-0000-000000000000"
		}

		companyID, err := uuid.Parse(companyIDStr.(string))
		if err != nil {
			utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Geçersiz firma", nil)
			c.Abort()
			return
		}

		var company models.Company
		if err := database.SystemDB.First(&company, companyID).Error; err != nil {
			utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Firma bulunamadı", nil)
			c.Abort()
			return
		}

		isTrialValid := company.SubscriptionStatus == "trial" && company.TrialEndsAt.After(time.Now())
		isActive := company.SubscriptionStatus == "active"

		if !isTrialValid && !isActive {
			utils.Err(c, http.StatusPaymentRequired, "SUBSCRIPTION_REQUIRED", "Aktif bir aboneliğiniz veya deneme süreniz bulunmamaktadır", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
