package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)

type AuditHandler struct{}

func NewAuditHandler() *AuditHandler {
	return &AuditHandler{}
}

func (h *AuditHandler) List(c *gin.Context) {
	tx := utils.GetDB(c.Request.Context(), database.DB)

	var list []models.AuditLog

	dbQuery := tx.Model(&models.AuditLog{})

	if moduleFilter := c.Query("module"); moduleFilter != "" {
		dbQuery = dbQuery.Where("module = ?", moduleFilter)
	}

	if recordIDFilter := c.Query("record_id"); recordIDFilter != "" {
		if rID, err := uuid.Parse(recordIDFilter); err == nil {
			dbQuery = dbQuery.Where("record_id = ?", rID)
		} else {
			utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz record_id formatı", nil)
			return
		}
	}

	if err := dbQuery.Order("created_at DESC").Find(&list).Error; err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, list)
}
