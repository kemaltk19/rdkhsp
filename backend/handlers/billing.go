package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"radikal-hesap/services"
	"radikal-hesap/utils"
)

type BillingHandler struct {
	service *services.BillingService
}

func NewBillingHandler(s *services.BillingService) *BillingHandler {
	return &BillingHandler{service: s}
}

func (h *BillingHandler) GetBillingStatus(c *gin.Context) {
	status, err := h.service.GetBillingStatus(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, status)
}

func (h *BillingHandler) GetPlans(c *gin.Context) {
	plans, err := h.service.GetPlans()
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, plans)
}

func (h *BillingHandler) Subscribe(c *gin.Context) {
	var in services.SubscribeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", err.Error(), nil)
		return
	}

	res, err := h.service.Subscribe(c.Request.Context(), in.PlanID, in.PeriodType)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", err.Error(), nil)
		return
	}
	utils.OK(c, res)
}

func (h *BillingHandler) ProcessWebhook(c *gin.Context) {
	// İmza doğrulaması için ham gövde gereklidir (yeniden serialize edilmiş
	// JSON imzayı bozar). Ham body'yi okuyup hem imza hem parse için kullan.
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "İstek gövdesi okunamadı", nil)
		return
	}

	var payload services.StripeWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz webhook verisi", nil)
		return
	}

	signature := c.GetHeader("Stripe-Signature")

	err = h.service.ProcessWebhook(rawBody, signature, payload)
	if err != nil {
		// İmza/event hatalarında 4xx; içeriden kaynaklı hatalarda 5xx.
		if errors.Is(err, services.ErrWebhookInvalidSignature) {
			utils.Err(c, http.StatusUnauthorized, "INVALID_SIGNATURE", "Webhook imzası doğrulanamadı", nil)
			return
		}
		if errors.Is(err, services.ErrWebhookMissingEventID) {
			utils.Err(c, http.StatusBadRequest, "MISSING_EVENT_ID", "event_id zorunludur", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, gin.H{"processed": true})
}

type RenewInput struct {
	PeriodType string `json:"period_type" binding:"required,oneof=monthly yearly"`
}

func (h *BillingHandler) Renew(c *gin.Context) {
	var in RenewInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", err.Error(), nil)
		return
	}

	res, err := h.service.Renew(c.Request.Context(), in.PeriodType)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", err.Error(), nil)
		return
	}
	utils.OK(c, res)
}

func (h *BillingHandler) GetTransactions(c *gin.Context) {
	txs, err := h.service.GetTransactions(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, txs)
}

