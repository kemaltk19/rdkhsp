package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"radikal-hesap/services"
	"radikal-hesap/utils"
)

type InvoiceHandler struct {
	svc *services.InvoiceService
}

func NewInvoiceHandler(svc *services.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{svc: svc}
}

func (h *InvoiceHandler) Create(c *gin.Context) {
	var in services.InvoiceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz fatura bilgileri", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.svc.Create(c.Request.Context(), in, userID)
	if err != nil {
		if errors.Is(err, services.ErrCariNotFoundForInvoice) {
			utils.Err(c, http.StatusNotFound, "CARI_NOT_FOUND", "Cari kart bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", invoiceSendErrorMessage(err), nil)
		return
	}

	utils.Created(c, res)
}

// invoiceSendErrorMessage, mail gönderim hatalarındaki teknik mesajları
// kullanıcının anlayacağı Türkçe karşılıklara çevirir.
func invoiceSendErrorMessage(err error) string {
	switch {
	case strings.Contains(err.Error(), "cari_email_not_found"):
		return "Cari hesabın e-posta adresi tanımlı değil, mail gönderilemedi"
	case strings.Contains(err.Error(), "email_send_failed"):
		return "Mail gönderilemedi, fatura kaydedilmedi: " + err.Error()
	case strings.Contains(err.Error(), "duplicate_invoice_number"):
		return "Bu fatura numarası zaten kullanılıyor, farklı bir numara girin"
	default:
		return err.Error()
	}
}

func (h *InvoiceHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz fatura ID", nil)
		return
	}

	var in services.InvoiceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz fatura bilgileri", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.svc.Update(c.Request.Context(), id, in, userID)
	if err != nil {
		if errors.Is(err, services.ErrInvoiceNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Fatura bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrInvoiceNotEditable) {
			utils.Err(c, http.StatusConflict, "NOT_EDITABLE", "Taslak dışındaki faturalar düzenlenemez", nil)
			return
		}
		if errors.Is(err, services.ErrCariNotFoundForInvoice) {
			utils.Err(c, http.StatusNotFound, "CARI_NOT_FOUND", "Cari kart bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", invoiceSendErrorMessage(err), nil)
		return
	}

	utils.OK(c, res)
}

type updateInvoiceStatusInput struct {
	Status    string          `json:"status" binding:"required,oneof=draft sent disputed partial paid canceled"`
	PaidTotal decimal.Decimal `json:"paid_total"`
}

func (h *InvoiceHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz fatura ID", nil)
		return
	}

	var in updateInvoiceStatusInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz durum bilgisi", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.svc.UpdateStatus(c.Request.Context(), id, in.Status, in.PaidTotal, userID)
	if err != nil {
		if errors.Is(err, services.ErrInvoiceNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Fatura bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusBadRequest, "INVALID_TRANSITION", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *InvoiceHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz fatura ID", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.svc.Delete(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, services.ErrInvoiceNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Fatura bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrInvoiceNotDeletable) {
			utils.Err(c, http.StatusConflict, "NOT_DELETABLE", "Taslak dışındaki faturalar silinemez. İptal etmeniz gerekir.", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"deleted": true})
}

func (h *InvoiceHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz fatura ID", nil)
		return
	}

	res, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrInvoiceNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Fatura bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *InvoiceHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	query := c.Query("q")
	sort := c.Query("sort")

	filters := make(map[string]string)
	if typeFilter := c.Query("type"); typeFilter != "" {
		filters["type"] = typeFilter
	}
	if statusFilter := c.Query("status"); statusFilter != "" {
		filters["status"] = statusFilter
	}
	if cariFilter := c.Query("cari_id"); cariFilter != "" {
		filters["cari_id"] = cariFilter
	}

	if limit > 100 {
		limit = 100
	}
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}

	res, total, err := h.svc.List(c.Request.Context(), page, limit, query, sort, filters)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.List(c, res, page, limit, total)
}

func (h *InvoiceHandler) Cancel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz fatura ID", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.svc.Cancel(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, services.ErrInvoiceNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Fatura bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrInvoiceAlreadyCanceled) {
			utils.Err(c, http.StatusConflict, "ALREADY_CANCELED", "Fatura zaten iptal edilmiş", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"canceled": true})
}

func (h *InvoiceHandler) Send(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz fatura ID", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.svc.SendInvoice(c.Request.Context(), id, userID)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", invoiceSendErrorMessage(err), nil)
		return
	}

	utils.OK(c, gin.H{"sent": true})
}

// BulkSend, seçilen her fatura için SendInvoice'u tek tek dener; bir faturanın
// başarısız olması (örn. cari emaili yok, durum uygun değil) diğerlerini durdurmaz.
// Sonunda gönderilen/başarısız id listesini özet olarak döner.
func (h *InvoiceHandler) BulkSend(c *gin.Context) {
	var in struct {
		IDs []uuid.UUID `json:"ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz fatura listesi", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	type failure struct {
		ID    uuid.UUID `json:"id"`
		Error string    `json:"error"`
	}
	sent := make([]uuid.UUID, 0, len(in.IDs))
	failed := make([]failure, 0)

	for _, id := range in.IDs {
		if err := h.svc.SendInvoice(c.Request.Context(), id, userID); err != nil {
			failed = append(failed, failure{ID: id, Error: err.Error()})
			continue
		}
		sent = append(sent, id)
	}

	utils.OK(c, gin.H{"sent": sent, "failed": failed})
}
