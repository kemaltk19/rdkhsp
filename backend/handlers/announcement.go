package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"radikal-hesap/services"
	"radikal-hesap/utils"
)

type AnnouncementHandler struct {
	svc *services.AnnouncementService
}

func NewAnnouncementHandler(svc *services.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{svc: svc}
}

type CreateAnnouncementInput struct {
	Title        string     `json:"title" binding:"required,max=255"`
	Body         string     `json:"body" binding:"required"`
	Category     string     `json:"category"`
	TargetPlanID *uuid.UUID `json:"target_plan_id"`
}

func (h *AnnouncementHandler) Create(c *gin.Context) {
	var in CreateAnnouncementInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", err.Error(), nil)
		return
	}

	creatorIDStr, _ := c.Get("user_id")
	creatorID, _ := uuid.Parse(creatorIDStr.(string))

	ann, err := h.svc.Create(c.Request.Context(), in.Title, in.Body, in.Category, in.TargetPlanID, creatorID)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	utils.OK(c, ann)
}

func (h *AnnouncementHandler) ListAll(c *gin.Context) {
	list, err := h.svc.ListAll(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, list)
}

func (h *AnnouncementHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"status": "deleted"})
}

func (h *AnnouncementHandler) ListForTenant(c *gin.Context) {
	compIDStr, _ := c.Get("company_id")
	companyID, _ := uuid.Parse(compIDStr.(string))

	list, err := h.svc.ListForTenant(c.Request.Context(), companyID)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, list)
}

// ── Kategori yönetimi ──

type CreateCategoryInput struct {
	Name string `json:"name" binding:"required,max=64"`
}

func (h *AnnouncementHandler) ListCategories(c *gin.Context) {
	list, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, list)
}

func (h *AnnouncementHandler) CreateCategory(c *gin.Context) {
	var in CreateCategoryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", err.Error(), nil)
		return
	}

	cat, err := h.svc.CreateCategory(c.Request.Context(), in.Name)
	if err != nil {
		switch err {
		case services.ErrCategoryExists:
			utils.Err(c, http.StatusBadRequest, "CATEGORY_EXISTS", "Bu kategori zaten mevcut", nil)
		case services.ErrCategoryNameEmpty:
			utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçerli bir kategori adı girin", nil)
		default:
			utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		}
		return
	}
	utils.OK(c, cat)
}

func (h *AnnouncementHandler) DeleteCategory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid id", nil)
		return
	}

	if err := h.svc.DeleteCategory(c.Request.Context(), id); err != nil {
		switch err {
		case services.ErrCategoryNotFound:
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Kategori bulunamadı", nil)
		case services.ErrCategoryProtected:
			utils.Err(c, http.StatusBadRequest, "CATEGORY_PROTECTED", "Varsayılan kategori silinemez", nil)
		default:
			utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		}
		return
	}
	utils.OK(c, gin.H{"status": "deleted"})
}
