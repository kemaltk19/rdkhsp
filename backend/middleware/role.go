package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"radikal-hesap/utils"
)

// RequireRole ensures that the user's role is one of the allowed roles.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			utils.Err(c, http.StatusForbidden, "FORBIDDEN", "Rol yetkisi bulunamadı", nil)
			c.Abort()
			return
		}

		role := roleVal.(string)
		for _, r := range allowedRoles {
			if r == role {
				c.Next()
				return
			}
		}

		utils.Err(c, http.StatusForbidden, "FORBIDDEN", "Bu işlem için yetkiniz bulunmamaktadır. Lütfen yöneticinize başvurun.", nil)
		c.Abort()
	}
}
