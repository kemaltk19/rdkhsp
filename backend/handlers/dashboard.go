package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"radikal-hesap/services"
	"radikal-hesap/utils"
)

type DashboardHandler struct {
	svc *services.DashboardService
}

func NewDashboardHandler(svc *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) GetStats(c *gin.Context) {
	res, err := h.svc.GetStats(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	utils.OK(c, res)
}
