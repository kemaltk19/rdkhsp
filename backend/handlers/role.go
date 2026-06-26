package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"radikal-hesap/services"
	"radikal-hesap/utils"
)

type RoleHandler struct {
	svc *services.RoleService
}

func NewRoleHandler(svc *services.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.svc.List(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, roles)
}

func (h *RoleHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid role id", nil)
		return
	}
	role, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrRoleNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Rol bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, role)
}

func (h *RoleHandler) Create(c *gin.Context) {
	var in services.RoleInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusBadRequest, "VALIDATION", err.Error(), nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	role, err := h.svc.Create(c.Request.Context(), in, userID)
	if err != nil {
		if errors.Is(err, services.ErrRoleInvalidModule) {
			utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz modül adı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": role})
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid role id", nil)
		return
	}
	var in services.RoleInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusBadRequest, "VALIDATION", err.Error(), nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	role, err := h.svc.Update(c.Request.Context(), id, in, userID)
	if err != nil {
		if errors.Is(err, services.ErrRoleNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Rol bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrRoleInvalidModule) {
			utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz modül adı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, role)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid role id", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	if err := h.svc.Delete(c.Request.Context(), id, userID); err != nil {
		if errors.Is(err, services.ErrRoleNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Rol bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrRoleInUse) {
			utils.Err(c, http.StatusBadRequest, "ROLE_IN_USE", "Bu role atanmış personeller var, önce onları başka bir role taşıyın", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	utils.OK(c, gin.H{"deleted": true})
}
