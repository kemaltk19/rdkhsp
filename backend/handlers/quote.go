package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"radikal-hesap/services"
	"radikal-hesap/utils"
)

// quoteSendErrorMessage, mail gönderim hatalarındaki teknik mesajları
// kullanıcının anlayacağı Türkçe karşılıklara çevirir.
func quoteSendErrorMessage(err error) string {
	switch {
	case strings.Contains(err.Error(), "cari_email_not_found"):
		return "Cari hesabın e-posta adresi tanımlı değil, mail gönderilemedi"
	case strings.Contains(err.Error(), "email gönderilemedi"):
		return "Mail gönderilemedi, teklif kaydedilmedi: " + err.Error()
	case strings.Contains(err.Error(), "duplicate_quote_number"):
		return "Bu teklif numarası zaten kullanılıyor, farklı bir numara girin"
	default:
		return err.Error()
	}
}

type QuoteHandler struct {
	svc *services.QuoteService
}

func NewQuoteHandler(svc *services.QuoteService) *QuoteHandler {
	return &QuoteHandler{svc: svc}
}

func (h *QuoteHandler) Create(c *gin.Context) {
	var in services.QuoteInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz teklif bilgileri", nil)
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
		if errors.Is(err, services.ErrCariNotFoundForQuote) {
			utils.Err(c, http.StatusNotFound, "CARI_NOT_FOUND", "Cari kart bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", quoteSendErrorMessage(err), nil)
		return
	}

	utils.Created(c, res)
}

func (h *QuoteHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz teklif ID", nil)
		return
	}

	var in services.QuoteInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz teklif bilgileri", nil)
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
		if errors.Is(err, services.ErrQuoteNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Teklif bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrQuoteNotEditable) {
			utils.Err(c, http.StatusConflict, "NOT_EDITABLE", "Kabul edilmiş veya faturaya dönüştürülmüş teklifler düzenlenemez", nil)
			return
		}
		if errors.Is(err, services.ErrCariNotFoundForQuote) {
			utils.Err(c, http.StatusNotFound, "CARI_NOT_FOUND", "Cari kart bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", quoteSendErrorMessage(err), nil)
		return
	}

	utils.OK(c, res)
}

func (h *QuoteHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz teklif ID", nil)
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
		if errors.Is(err, services.ErrQuoteNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Teklif bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrQuoteNotDeletable) {
			utils.Err(c, http.StatusConflict, "NOT_DELETABLE", "Kabul edilmiş veya faturaya dönüştürülmüş teklifler silinemez", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"deleted": true})
}

func (h *QuoteHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz teklif ID", nil)
		return
	}

	res, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrQuoteNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Teklif bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *QuoteHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	query := c.Query("q")
	sort := c.Query("sort")

	filters := make(map[string]string)
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

type updateStatusInput struct {
	Status string `json:"status" binding:"required,oneof=draft sent accepted rejected expired"`
}

func (h *QuoteHandler) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz teklif ID", nil)
		return
	}

	var in updateStatusInput
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

	err = h.svc.UpdateStatus(c.Request.Context(), id, in.Status, userID)
	if err != nil {
		if errors.Is(err, services.ErrQuoteNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Teklif bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"updated": true})
}

func (h *QuoteHandler) Convert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz teklif ID", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.svc.Convert(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, services.ErrQuoteNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Teklif bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrQuoteAlreadyConverted) {
			utils.Err(c, http.StatusConflict, "ALREADY_CONVERTED", "Teklif zaten faturaya dönüştürülmüş", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *QuoteHandler) Send(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz teklif ID", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.svc.SendQuote(c.Request.Context(), id, userID)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", quoteSendErrorMessage(err), nil)
		return
	}

	utils.OK(c, gin.H{"sent": true})
}

// BulkSend, seçilen her teklif için SendQuote'u tek tek dener; bir teklifin
// başarısız olması (örn. cari emaili yok, durum uygun değil) diğerlerini durdurmaz.
// Sonunda gönderilen/başarısız id listesini özet olarak döner.
func (h *QuoteHandler) BulkSend(c *gin.Context) {
	var in struct {
		IDs []uuid.UUID `json:"ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz teklif listesi", nil)
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
		if err := h.svc.SendQuote(c.Request.Context(), id, userID); err != nil {
			failed = append(failed, failure{ID: id, Error: err.Error()})
			continue
		}
		sent = append(sent, id)
	}

	utils.OK(c, gin.H{"sent": sent, "failed": failed})
}
