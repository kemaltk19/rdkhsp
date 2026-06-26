package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"radikal-hesap/services"
	"radikal-hesap/utils"
)

type CariHandler struct {
	svc *services.CariService
}

func NewCariHandler(svc *services.CariService) *CariHandler {
	return &CariHandler{svc: svc}
}

func (h *CariHandler) Create(c *gin.Context) {
	var in services.CariInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz cari bilgileri", nil)
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
		if errors.Is(err, services.ErrDuplicateCariCode) {
			utils.Err(c, http.StatusConflict, "DUPLICATE_CODE", "Bu cari kodu zaten kullanımda", nil)
			return
		}
		if errors.Is(err, services.ErrDuplicateTaxNumber) {
			utils.Err(c, http.StatusConflict, "DUPLICATE_TAX_NUMBER", "Bu TCKN veya Vergi No sistemde mevcut", nil)
			return
		}
		if errors.Is(err, services.ErrDuplicatePhone) {
			utils.Err(c, http.StatusConflict, "DUPLICATE_PHONE", "Bu telefon numarası sistemde mevcut", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.Created(c, res)
}

func (h *CariHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz cari ID", nil)
		return
	}

	var in services.CariInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz form verisi: "+err.Error(), nil)
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
		if errors.Is(err, services.ErrCariNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Cari kart bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrDuplicateTaxNumber) {
			utils.Err(c, http.StatusConflict, "DUPLICATE_TAX_NUMBER", "Bu TCKN veya Vergi No sistemde mevcut", nil)
			return
		}
		if errors.Is(err, services.ErrDuplicatePhone) {
			utils.Err(c, http.StatusConflict, "DUPLICATE_PHONE", "Bu telefon numarası sistemde mevcut", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *CariHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz cari ID", nil)
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
		if errors.Is(err, services.ErrCariNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Cari kart bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrCariHasTransactions) {
			utils.Err(c, http.StatusConflict, "CONFLICT", "Hareket görmüş cari kart silinemez", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"deleted": true})
}

func (h *CariHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz cari ID", nil)
		return
	}

	res, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrCariNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Cari kart bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *CariHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	query := c.Query("q")
	sort := c.Query("sort")

	// Collect filters
	filters := make(map[string]string)
	if tFilter := c.Query("type"); tFilter != "" {
		filters["type"] = tFilter
	}
	if gFilter := c.Query("group"); gFilter != "" {
		filters["group"] = gFilter
	}
	if activeFilter := c.Query("is_active"); activeFilter != "" {
		filters["is_active"] = activeFilter
	}

	// Limit bounds
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

func (h *CariHandler) GetTransactions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz cari ID", nil)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit > 100 {
		limit = 100
	}
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}

	res, total, err := h.svc.GetTransactions(c.Request.Context(), id, page, limit)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.List(c, res, page, limit, total)
}

func (h *CariHandler) GetSummary(c *gin.Context) {
	res, err := h.svc.GetSummary(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *CariHandler) GetCariFinancialSummary(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz cari ID", nil)
		return
	}

	res, err := h.svc.GetFinancialSummary(c.Request.Context(), id)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *CariHandler) GetNextCode(c *gin.Context) {
	res, err := h.svc.GetNextCode(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *CariHandler) AddPerson(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz cari ID", nil)
		return
	}

	var req struct {
		Name  string `json:"name" binding:"required"`
		Title string `json:"title"`
		Phone string `json:"phone"`
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Kişi adı gereklidir", nil)
		return
	}

	res, err := h.svc.AddPerson(c.Request.Context(), id, req.Name, req.Title, req.Phone, req.Email)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.Created(c, res)
}

func (h *CariHandler) RemovePerson(c *gin.Context) {
	personIdStr := c.Param("person_id")
	personId, err := uuid.Parse(personIdStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz kişi ID", nil)
		return
	}

	err = h.svc.RemovePerson(c.Request.Context(), personId)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"deleted": true})
}

func (h *CariHandler) UpdatePerson(c *gin.Context) {
	personIdStr := c.Param("person_id")
	personId, err := uuid.Parse(personIdStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz kişi ID", nil)
		return
	}

	var req struct {
		Name  string `json:"name" binding:"required"`
		Title string `json:"title"`
		Phone string `json:"phone"`
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Kişi adı gereklidir", nil)
		return
	}

	res, err := h.svc.UpdatePerson(c.Request.Context(), personId, req.Name, req.Title, req.Phone, req.Email)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}
