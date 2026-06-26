package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/services"
	"radikal-hesap/utils"
)

type PublicQuoteHandler struct {
	svc *services.QuoteService
}

func NewPublicQuoteHandler(svc *services.QuoteService) *PublicQuoteHandler {
	return &PublicQuoteHandler{svc: svc}
}

func (h *PublicQuoteHandler) GetByToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Token eksik", nil)
		return
	}

	// Public token erişimi tenant context'i taşımaz (ziyaretçi login değildir),
	// bu yüzden RLS uygulanan database.DB'de satır görünmez; RLS bypass eden
	// SystemDB kullanılır. Token zaten gizli/rastgele olduğu için güvenlidir.

	// Automatically transition if expired
	database.SystemDB.Model(&models.Quote{}).Where("public_token = ? AND status = 'sent' AND expiry_date < ?", token, time.Now()).Update("status", "expired")

	var quote models.Quote
	// Eager load items and cari to display details
	if err := database.SystemDB.Preload("Items").First(&quote, "public_token = ?", token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Teklif bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	var cari models.Cari
	if err := database.SystemDB.First(&cari, "id = ?", quote.CariID).Error; err == nil {
		// Loaded safely
	}

	// Müşteri için güvenli özet dönüş
	res := gin.H{
		"id":             quote.ID,
		"number":         quote.Number,
		"date":           quote.Date,
		"expiry_date":    quote.ExpiryDate,
		"currency":       quote.Currency,
		"total":          quote.Total,
		"subtotal":       quote.Subtotal,
		"tax_total":      quote.TaxTotal,
		"discount_total": quote.DiscountTotal,
		"status":         quote.Status,
		"note":           quote.Note,
		"reject_note":    quote.RejectNote,
		"items":          quote.Items,
		"cari": gin.H{
			"name":  cari.Name,
			"email": cari.Email,
			"phone": cari.Phone,
		},
	}

	utils.OK(c, res)
}

func (h *PublicQuoteHandler) AcceptPublic(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Token eksik", nil)
		return
	}

	invoice, err := h.svc.AcceptPublic(c.Request.Context(), token)
	if err != nil {
		if err.Error() == "bu teklif daha önce yanıtlanmış veya süresi dolmuş" {
			utils.Err(c, http.StatusConflict, "ALREADY_RESPONDED", err.Error(), nil)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Teklif bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	// Kabul sonucu oluşan fatura bilgisini döndür
	utils.OK(c, gin.H{
		"accepted": true,
		"invoice": gin.H{
			"id":     invoice.ID,
			"number": invoice.Number,
		},
	})
}

type QuoteRejectInput struct {
	Note string `json:"note" binding:"required,max=2000"`
}

func (h *PublicQuoteHandler) RejectPublic(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Token eksik", nil)
		return
	}

	var in QuoteRejectInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Red nedeni zorunludur", nil)
		return
	}

	err := h.svc.RejectPublic(c.Request.Context(), token, in.Note)
	if err != nil {
		if err.Error() == "bu teklif daha önce yanıtlanmış veya süresi dolmuş" {
			utils.Err(c, http.StatusConflict, "ALREADY_RESPONDED", err.Error(), nil)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Teklif bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"rejected": true})
}
