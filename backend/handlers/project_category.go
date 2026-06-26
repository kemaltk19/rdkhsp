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

type ProjectCategoryHandler struct {
	svc *services.ProjectCategoryService
}

func NewProjectCategoryHandler(svc *services.ProjectCategoryService) *ProjectCategoryHandler {
	return &ProjectCategoryHandler{svc: svc}
}

func (h *ProjectCategoryHandler) Create(c *gin.Context) {
	var in services.ProjectCategoryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz proje kategorisi bilgileri", nil)
		return
	}

	res, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.Created(c, res)
}

func (h *ProjectCategoryHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz proje kategorisi ID", nil)
		return
	}

	var in services.ProjectCategoryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz proje kategorisi bilgileri", nil)
		return
	}

	res, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		if errors.Is(err, services.ErrProjectCategoryNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Proje kategorisi bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *ProjectCategoryHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz proje kategorisi ID", nil)
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, services.ErrProjectCategoryNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Proje kategorisi bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"deleted": true})
}

func (h *ProjectCategoryHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz proje kategorisi ID", nil)
		return
	}

	res, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrProjectCategoryNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Proje kategorisi bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *ProjectCategoryHandler) List(c *gin.Context) {
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	query := c.Query("query")
	sort := c.Query("sort")

	categories, total, err := h.svc.List(c.Request.Context(), page, limit, query, sort)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.List(c, categories, page, limit, total)
}
