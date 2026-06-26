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

type NotificationHandler struct {
	svc *services.NotificationService
}

func NewNotificationHandler(svc *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	notifications, total, err := h.svc.List(c.Request.Context(), page, limit)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{
		"data": notifications,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz bildirim ID", nil)
		return
	}

	err = h.svc.MarkAsRead(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrNotificationNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"success": true})
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	err := h.svc.MarkAllAsRead(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"success": true})
}
