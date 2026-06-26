package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"radikal-hesap/services"
	"radikal-hesap/utils"
)

type ReportHandler struct {
	svc *services.ReportService
}

func NewReportHandler(svc *services.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

func (h *ReportHandler) GetReport(c *gin.Context) {
	reportType := c.Param("type")

	filters := make(map[string]string)
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		filters["date_from"] = dateFrom
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		filters["date_to"] = dateTo
	}
	if cariID := c.Query("cari_id"); cariID != "" {
		filters["cari_id"] = cariID
	}

	res, err := h.svc.GetReport(c.Request.Context(), reportType, filters)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}
