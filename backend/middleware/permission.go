package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)

// RequireModulePermission, "personel" rolündeki kullanıcıların modül+aksiyon
// bazlı izinlerini RolePermission tablosundan kontrol eder.
// admin/superadmin her zaman geçer (implicit full-access); cari zaten ayrı
// portalda çalıştığı için bu middleware'in kapsadığı route'lara girmez.
// Ayrıca şirketin plan özelliklerini ve aktif modüllerini denetler.
func RequireModulePermission(module string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, _ := c.Get("role")
		role, _ := roleVal.(string)

		if role == "admin" || role == "superadmin" {
			c.Next()
			return
		}

		if role != "personel" {
			utils.Err(c, http.StatusForbidden, "FORBIDDEN", "Bu işlem için yetkiniz bulunmamaktadır. Lütfen yöneticinize başvurun.", nil)
			c.Abort()
			return
		}

		companyIDStr, exists := c.Get("company_id")
		if !exists || companyIDStr == "" {
			utils.Err(c, http.StatusForbidden, "FORBIDDEN", "Firma bilgisi bulunamadı", nil)
			c.Abort()
			return
		}

		companyID, err := uuid.Parse(companyIDStr.(string))
		if err != nil {
			utils.Err(c, http.StatusForbidden, "FORBIDDEN", "Geçersiz firma", nil)
			c.Abort()
			return
		}

		// Plan features & enabled modules check
		var company models.Company
		if err := database.SystemDB.First(&company, "id = ?", companyID).Error; err == nil {
			// Enforce plan features
			if company.PlanID != nil {
				var plan models.Plan
				if err := database.SystemDB.First(&plan, "id = ?", *company.PlanID).Error; err == nil {
					if plan.Features != "" {
						var features []string
						if err := json.Unmarshal([]byte(plan.Features), &features); err == nil {
							hasFeature := false
							for _, f := range features {
								if f == module {
									hasFeature = true
									break
								}
							}
							if !hasFeature {
								utils.Err(c, http.StatusForbidden, "FORBIDDEN", "Bu modül paketinizde bulunmamaktadır.", nil)
								c.Abort()
								return
							}
						}
					}
				}
			}

			// Enforce enabled modules set by admin
			if company.EnabledModules != nil && *company.EnabledModules != "" {
				var enabled []string
				if err := json.Unmarshal([]byte(*company.EnabledModules), &enabled); err == nil {
					isModuleEnabled := false
					for _, m := range enabled {
						if m == module {
							isModuleEnabled = true
							break
						}
					}
					if !isModuleEnabled {
						utils.Err(c, http.StatusForbidden, "FORBIDDEN", "Bu modül firma yöneticisi tarafından devre dışı bırakılmıştır.", nil)
						c.Abort()
						return
					}
				}
			}
		}

		userIDStr, exists := c.Get("user_id")
		if !exists {
			utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekmektedir", nil)
			c.Abort()
			return
		}
		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Geçersiz oturum", nil)
			c.Abort()
			return
		}

		var user models.User
		if err := database.SystemDB.Select("role_id").First(&user, "id = ?", userID).Error; err != nil || user.RoleID == nil {
			utils.Err(c, http.StatusForbidden, "FORBIDDEN", "Herhangi bir role atanmamışsınız, lütfen yöneticinizle iletişime geçin", nil)
			c.Abort()
			return
		}

		var perm models.RolePermission
		if err := database.SystemDB.
			Where("role_id = ? AND module = ?", *user.RoleID, module).
			First(&perm).Error; err != nil {
			utils.Err(c, http.StatusForbidden, "FORBIDDEN", "Bu modül için yetkiniz bulunmamaktadır. Lütfen yöneticinize başvurun.", nil)
			c.Abort()
			return
		}

		allowed := false
		switch action {
		case "create":
			allowed = perm.CanCreate
		case "read":
			allowed = perm.CanRead
		case "update":
			allowed = perm.CanUpdate
		case "delete":
			allowed = perm.CanDelete
		}

		if !allowed {
			utils.Err(c, http.StatusForbidden, "FORBIDDEN", "Bu işlem için yetkiniz bulunmamaktadır. Lütfen yöneticinize başvurun.", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
