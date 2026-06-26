package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"radikal-hesap/services"
	"radikal-hesap/utils"
)

type SettingsHandler struct {
	service *services.SettingsService
}

func NewSettingsHandler(s *services.SettingsService) *SettingsHandler {
	return &SettingsHandler{service: s}
}

func (h *SettingsHandler) GetCompanyProfile(c *gin.Context) {
	comp, err := h.service.GetCompanyProfile(c.Request.Context())
	if err != nil {
		if err == services.ErrCompanyNotFound {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "company_not_found", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, comp)
}

func (h *SettingsHandler) UpdateCompanyProfile(c *gin.Context) {
	var in services.UpdateCompanyInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", err.Error(), nil)
		return
	}

	comp, err := h.service.UpdateCompanyProfile(c.Request.Context(), in)
	if err != nil {
		if err == services.ErrCompanyNotFound {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "company_not_found", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, comp)
}

func (h *SettingsHandler) GetSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "key_required", nil)
		return
	}

	setting, err := h.service.GetSetting(c.Request.Context(), key)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	if setting == nil {
		utils.OK(c, gin.H{"key": key, "value": ""})
		return
	}

	utils.OK(c, setting)
}

func (h *SettingsHandler) SaveSetting(c *gin.Context) {
	var in services.SaveSettingInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", err.Error(), nil)
		return
	}

	setting, err := h.service.SaveSetting(c.Request.Context(), in)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, setting)
}

func (h *SettingsHandler) ListSettings(c *gin.Context) {
	category := c.Query("category")
	settings, err := h.service.ListSettings(c.Request.Context(), category)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, settings)
}

func (h *SettingsHandler) UpdateEnabledModules(c *gin.Context) {
	var in struct {
		EnabledModules []string `json:"enabled_modules" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", err.Error(), nil)
		return
	}

	comp, err := h.service.UpdateEnabledModules(c.Request.Context(), in.EnabledModules)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, comp)
}
