package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"radikal-hesap/services"
	"radikal-hesap/utils"
)

type ProductHandler struct {
	svc *services.StockService
}

func NewProductHandler(svc *services.StockService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// ----------------------------------------------------
// Product Category Handlers
// ----------------------------------------------------

func (h *ProductHandler) CreateCategory(c *gin.Context) {
	var in struct {
		Name           string          `json:"name" binding:"required"`
		DefaultKDVRate decimal.Decimal `json:"default_kdv_rate"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Kategori adı zorunludur", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Yetkisiz erişim", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.svc.CreateCategory(c.Request.Context(), in.Name, in.DefaultKDVRate, userID)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.Created(c, res)
}

func (h *ProductHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz kategori ID", nil)
		return
	}

	var in struct {
		Name           string          `json:"name" binding:"required"`
		DefaultKDVRate decimal.Decimal `json:"default_kdv_rate"`
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

	res, err := h.svc.UpdateCategory(c.Request.Context(), id, in.Name, in.DefaultKDVRate, userID)
	if err != nil {
		if errors.Is(err, services.ErrProductCategoryNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Kategori bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *ProductHandler) DeleteCategory(c *gin.Context) {
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

	err = h.svc.DeleteCategory(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, services.ErrCategoryInUse) {
			utils.Err(c, http.StatusConflict, "CONFLICT", "Kullanımda olan kategori silinemez", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"deleted": true})
}

func (h *ProductHandler) ListCategories(c *gin.Context) {
	res, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	utils.List(c, res, 1, len(res), int64(len(res)))
}

// ----------------------------------------------------
// Warehouse Handlers
// ----------------------------------------------------

func (h *ProductHandler) CreateWarehouse(c *gin.Context) {
	var in struct {
		Name      string `json:"name" binding:"required"`
		Address   string `json:"address"`
		IsDefault bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Depo adı zorunludur", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.svc.CreateWarehouse(c.Request.Context(), in.Name, in.Address, in.IsDefault, userID)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.Created(c, res)
}

func (h *ProductHandler) UpdateWarehouse(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz depo ID", nil)
		return
	}

	var in struct {
		Name      string `json:"name" binding:"required"`
		Address   string `json:"address"`
		IsDefault bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Depo adı zorunludur", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.svc.UpdateWarehouse(c.Request.Context(), id, in.Name, in.Address, in.IsDefault, userID)
	if err != nil {
		if errors.Is(err, services.ErrWarehouseNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Depo bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *ProductHandler) DeleteWarehouse(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz depo ID", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.svc.DeleteWarehouse(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, services.ErrWarehouseInUse) {
			utils.Err(c, http.StatusConflict, "CONFLICT", "Hareket görmüş depo silinemez", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"deleted": true})
}

func (h *ProductHandler) ListWarehouses(c *gin.Context) {
	res, err := h.svc.ListWarehouses(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	utils.List(c, res, 1, len(res), int64(len(res)))
}

// ----------------------------------------------------
// Product Handlers
// ----------------------------------------------------

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var in services.ProductInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz ürün verileri: "+err.Error(), nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.svc.CreateProduct(c.Request.Context(), in, userID)
	if err != nil {
		if errors.Is(err, services.ErrProductCategoryNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Kategori bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrDuplicateProductCode) {
			utils.Err(c, http.StatusConflict, "DUPLICATE_CODE", "Bu ürün kodu zaten kullanımda", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.Created(c, res)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz ürün ID", nil)
		return
	}

	var in services.ProductInput
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

	res, err := h.svc.UpdateProduct(c.Request.Context(), id, in, userID)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Ürün bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrProductCategoryNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Kategori bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrDuplicateProductCode) {
			utils.Err(c, http.StatusConflict, "DUPLICATE_CODE", "Bu ürün kodu zaten kullanımda", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz ürün ID", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.svc.DeleteProduct(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Ürün bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrProductInUse) {
			utils.Err(c, http.StatusConflict, "CONFLICT", "Hareket görmüş veya faturada kullanılmış ürün silinemez", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"deleted": true})
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz ürün ID", nil)
		return
	}

	res, err := h.svc.GetProductByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Ürün bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
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
	if typeFilter := c.Query("type"); typeFilter != "" {
		filters["type"] = typeFilter
	}
	if categoryID := c.Query("category_id"); categoryID != "" {
		filters["category_id"] = categoryID
	}

	res, total, err := h.svc.ListProducts(c.Request.Context(), page, limit, query, sort, filters)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.List(c, res, page, limit, total)
}

// ----------------------------------------------------
// Stock Movement Handlers
// ----------------------------------------------------

func (h *ProductHandler) ManualAdjustment(c *gin.Context) {
	var in services.StockMovementInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz veri: "+err.Error(), nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.svc.ManualAdjustment(c.Request.Context(), in, userID)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Ürün bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrWarehouseNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Depo bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.Created(c, res)
}

func (h *ProductHandler) ListMovements(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz ürün ID", nil)
		return
	}

	res, err := h.svc.ListMovements(c.Request.Context(), id)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.List(c, res, 1, len(res), int64(len(res)))
}

func (h *ProductHandler) GetCriticalStock(c *gin.Context) {
	items, err := h.svc.GetCriticalStockProducts(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	utils.OK(c, items)
}

func (h *ProductHandler) GetNextCode(c *gin.Context) {
	res, err := h.svc.GetNextCode(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	utils.OK(c, res)
}
