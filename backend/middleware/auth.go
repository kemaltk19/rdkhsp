package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"radikal-hesap/utils"
)

// AuthMiddleware verifies the JWT token from cookies or Authorization header.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie("at")
		if err != nil || tokenStr == "" {
			// Fallback to Authorization Header
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenStr == "" {
			utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekmektedir", nil)
			c.Abort()
			return
		}

		secretStr := os.Getenv("JWT_SECRET")
		if secretStr == "" {
			utils.Err(c, http.StatusInternalServerError, "INTERNAL", "Sistem anahtarı eksik", nil)
			c.Abort()
			return
		}

		claims, err := utils.ParseAccess([]byte(secretStr), tokenStr)
		if err != nil {
			utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Geçersiz veya süresi dolmuş oturum", nil)
			c.Abort()
			return
		}

		// Store claims in gin context
		c.Set("user_id", claims.UserID.String())
		c.Set("company_id", claims.CompanyID.String())
		c.Set("role", claims.Role)

		// Store in request context too
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "user_id", claims.UserID.String())
		ctx = context.WithValue(ctx, "company_id", claims.CompanyID.String())
		ctx = context.WithValue(ctx, "role", claims.Role)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
