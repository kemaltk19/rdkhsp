package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *HealthHandler) HealthDB(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db yok"})
		return
	}
	sqlDB, err := h.db.DB()
	if err == nil {
		err = sqlDB.Ping()
	}
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db hatasi", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "db ok"})
}
