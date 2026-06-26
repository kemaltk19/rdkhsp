package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"radikal-hesap/services"
	"radikal-hesap/utils"
)

type EmployeeHandler struct {
	service *services.EmployeeService
}

func NewEmployeeHandler(s *services.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{service: s}
}

func (h *EmployeeHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	sort := c.Query("sort")

	var isActive *bool
	if activeStr := c.Query("is_active"); activeStr != "" {
		actBool := activeStr == "true"
		isActive = &actBool
	}

	employees, total, err := h.service.List(c.Request.Context(), page, limit, search, sort, isActive)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	utils.List(c, employees, page, limit, total)
}

func (h *EmployeeHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid_id", nil)
		return
	}

	emp, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == services.ErrEmployeeNotFound {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "employee_not_found", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	utils.OK(c, emp)
}

func (h *EmployeeHandler) Create(c *gin.Context) {
	var in services.CreateEmployeeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", err.Error(), nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	emp, err := h.service.Create(c.Request.Context(), in, userID)
	if err != nil {
		if err == services.ErrEmployeeEmailExists {
			utils.Err(c, http.StatusBadRequest, "EMAIL_ALREADY_REGISTERED", "email_already_registered", nil)
			return
		}
		if err == services.ErrRoleNotFoundOrForbidden {
			utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Seçilen rol bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	utils.Created(c, emp)
}

func (h *EmployeeHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid_id", nil)
		return
	}

	var in services.UpdateEmployeeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", err.Error(), nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	emp, err := h.service.Update(c.Request.Context(), id, in, userID)
	if err != nil {
		if err == services.ErrEmployeeNotFound {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "employee_not_found", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	utils.OK(c, emp)
}

func (h *EmployeeHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "invalid_id", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.service.Delete(c.Request.Context(), id, userID)
	if err != nil {
		if err == services.ErrEmployeeNotFound {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "employee_not_found", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"deleted": true})
}
