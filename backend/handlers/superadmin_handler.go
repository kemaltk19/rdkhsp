package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"radikal-hesap/models"
	"radikal-hesap/services"
	"radikal-hesap/utils"
)

type SuperadminHandler struct {
	svc *services.SuperadminService
}

func NewSuperadminHandler(svc *services.SuperadminService) *SuperadminHandler {
	return &SuperadminHandler{svc: svc}
}

// -----------------------------------------------------------------------------
// Dashboard Stats
// -----------------------------------------------------------------------------

func (h *SuperadminHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.svc.GetDashboardStats(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, stats)
}

// -----------------------------------------------------------------------------
// Company Management
// -----------------------------------------------------------------------------

func (h *SuperadminHandler) GetCompanies(c *gin.Context) {
	companies, err := h.svc.GetCompanies(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, companies)
}

func (h *SuperadminHandler) ToggleCompanyStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid company id", nil)
		return
	}

	var req struct {
		Action string `json:"action" binding:"required"` // "suspend" or "activate"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid payload", nil)
		return
	}

	if err := h.svc.ToggleCompanyStatus(c.Request.Context(), id, req.Action); err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"status": "success", "action": req.Action})
}

func (h *SuperadminHandler) CreateCompany(c *gin.Context) {
	var in services.CreateCompanyInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", err.Error(), nil)
		return
	}

	comp, err := h.svc.CreateCompany(c.Request.Context(), in)
	if err != nil {
		if err.Error() == "email_exists" {
			utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Bu e-posta adresiyle kayıtlı bir kullanıcı zaten var.", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	utils.OK(c, comp)
}

func (h *SuperadminHandler) UpdateCompany(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid company id", nil)
		return
	}

	var in services.SuperadminUpdateCompanyInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", err.Error(), nil)
		return
	}

	comp, err := h.svc.UpdateCompany(c.Request.Context(), id, in)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	utils.OK(c, comp)
}

func (h *SuperadminHandler) DeleteCompany(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid company id", nil)
		return
	}

	if err := h.svc.DeleteCompany(c.Request.Context(), id); err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"status": "deleted"})
}

// -----------------------------------------------------------------------------
// Plan Management
// -----------------------------------------------------------------------------

func (h *SuperadminHandler) GetPlans(c *gin.Context) {
	plans, err := h.svc.GetPlans(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, plans)
}

func (h *SuperadminHandler) CreatePlan(c *gin.Context) {
	var plan models.Plan
	if err := c.ShouldBindJSON(&plan); err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid data", nil)
		return
	}

	if err := h.svc.CreatePlan(c.Request.Context(), &plan); err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": plan})
}

func (h *SuperadminHandler) UpdatePlan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid plan id", nil)
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid data", nil)
		return
	}

	if err := h.svc.UpdatePlan(c.Request.Context(), id, data); err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"status": "updated"})
}

func (h *SuperadminHandler) DeletePlan(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid plan id", nil)
		return
	}

	if err := h.svc.DeletePlan(c.Request.Context(), id); err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"status": "deactivated"})
}

// -----------------------------------------------------------------------------
// Email (SMTP) Settings — platform-wide
// -----------------------------------------------------------------------------

func (h *SuperadminHandler) GetEmailSettings(c *gin.Context) {
	es, err := h.svc.GetEmailSettings(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	// Password is never returned; expose only whether one is set.
	utils.OK(c, gin.H{
		"host":         es.Host,
		"port":         es.Port,
		"username":     es.Username,
		"from_email":   es.FromEmail,
		"from_name":    es.FromName,
		"enabled":      es.Enabled,
		"has_password": es.PasswordEnc != "",
	})
}

func (h *SuperadminHandler) UpdateEmailSettings(c *gin.Context) {
	var in services.EmailSettingsInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid data", nil)
		return
	}
	es, err := h.svc.SaveEmailSettings(c.Request.Context(), in)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, gin.H{"enabled": es.Enabled, "has_password": es.PasswordEnc != ""})
}

func (h *SuperadminHandler) TestEmailSettings(c *gin.Context) {
	var req struct {
		To string `json:"to" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "geçerli bir e-posta adresi gir", nil)
		return
	}
	if err := h.svc.SendTestEmail(c.Request.Context(), req.To); err != nil {
		utils.Err(c, http.StatusInternalServerError, "MAIL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, gin.H{"sent": true})
}
