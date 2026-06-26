package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"radikal-hesap/services"
	"radikal-hesap/utils"
)

type CurrencyHandler struct {
	svc *services.CurrencyService
}

func NewCurrencyHandler(svc *services.CurrencyService) *CurrencyHandler {
	return &CurrencyHandler{svc: svc}
}

func (h *CurrencyHandler) List(c *gin.Context) {
	currencies, err := h.svc.List(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, currencies)
}

func (h *CurrencyHandler) Create(c *gin.Context) {
	var in services.CurrencyInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusBadRequest, "VALIDATION", err.Error(), nil)
		return
	}
	currency, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": currency})
}

func (h *CurrencyHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid currency id", nil)
		return
	}
	var in services.CurrencyInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusBadRequest, "VALIDATION", err.Error(), nil)
		return
	}
	currency, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		if errors.Is(err, services.ErrCurrencyNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Para birimi bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, currency)
}

func (h *CurrencyHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid currency id", nil)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, services.ErrCurrencyNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Para birimi bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrCurrencyDefaultDelete) {
			utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Varsayılan para birimi silinemez", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, gin.H{"deleted": true})
}
