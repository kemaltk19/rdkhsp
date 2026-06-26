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

type ExpenseHandler struct {
	svc    *services.ExpenseService
	catSvc *services.ExpenseCategoryService
}

func NewExpenseHandler(svc *services.ExpenseService, catSvc *services.ExpenseCategoryService) *ExpenseHandler {
	return &ExpenseHandler{svc: svc, catSvc: catSvc}
}

// Category Handlers

func (h *ExpenseHandler) CreateCategory(c *gin.Context) {
	var in struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Kategori adı zorunludur", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.catSvc.CreateCategory(c.Request.Context(), in.Name, userID)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.Created(c, res)
}

func (h *ExpenseHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz kategori ID", nil)
		return
	}

	var in struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Kategori adı zorunludur", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.catSvc.UpdateCategory(c.Request.Context(), id, in.Name, userID)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *ExpenseHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz kategori ID", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.catSvc.DeleteCategory(c.Request.Context(), id, userID)
	if err != nil {
		if err.Error() == "category_in_use_cannot_delete" {
			utils.Err(c, http.StatusConflict, "CONFLICT", "Kullanımda olan gider kategorisi silinemez", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"deleted": true})
}

func (h *ExpenseHandler) ListCategories(c *gin.Context) {
	res, err := h.catSvc.ListCategories(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	utils.List(c, res, 1, len(res), int64(len(res)))
}

// Expense Handlers

func (h *ExpenseHandler) Create(c *gin.Context) {
	var in services.ExpenseInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz gider verileri: "+err.Error(), nil)
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
		if errors.Is(err, services.ErrExpenseCategoryNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Gider kategorisi bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrInvalidAccountForExpense) {
			utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz ödeme hesabı", nil)
			return
		}
		if errors.Is(err, services.ErrCariNotFoundForExpense) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Cari kart bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.Created(c, res)
}

func (h *ExpenseHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz gider ID", nil)
		return
	}

	var in services.ExpenseInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz form verisi", nil)
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
		if errors.Is(err, services.ErrExpenseNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Gider kaydı bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrExpenseAlreadyCanceled) {
			utils.Err(c, http.StatusConflict, "ALREADY_CANCELED", "İptal edilmiş gider kaydı güncellenemez", nil)
			return
		}
		if errors.Is(err, services.ErrExpenseCategoryNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Gider kategorisi bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrInvalidAccountForExpense) {
			utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz ödeme hesabı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *ExpenseHandler) Cancel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz gider ID", nil)
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
		if errors.Is(err, services.ErrExpenseNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Gider kaydı bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrExpenseAlreadyCanceled) {
			utils.Err(c, http.StatusConflict, "ALREADY_CANCELED", "Gider kaydı zaten iptal edilmiş", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"canceled": true})
}

func (h *ExpenseHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz gider ID", nil)
		return
	}

	res, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrExpenseNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Gider kaydı bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *ExpenseHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	query := c.Query("q")
	sort := c.Query("sort")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	filters := make(map[string]string)
	if categoryID := c.Query("category_id"); categoryID != "" {
		filters["category_id"] = categoryID
	}
	if cariID := c.Query("cari_id"); cariID != "" {
		filters["cari_id"] = cariID
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	res, total, err := h.svc.List(c.Request.Context(), page, limit, query, sort, filters)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.List(c, res, page, limit, total)
}

func (h *ExpenseHandler) GetRepeatAnalysis(c *gin.Context) {
	results, err := h.svc.GetRepeatAnalysis(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	utils.OK(c, results)
}
