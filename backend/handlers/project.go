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

type ProjectHandler struct {
	svc *services.ProjectService
}

func NewProjectHandler(svc *services.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

func (h *ProjectHandler) Create(c *gin.Context) {
	var in services.ProjectInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz proje bilgileri", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.svc.Create(c.Request.Context(), in, userID)
	if err != nil {
		if errors.Is(err, services.ErrCariNotFoundForProject) {
			utils.Err(c, http.StatusNotFound, "CARI_NOT_FOUND", "Cari kart bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.Created(c, res)
}

func (h *ProjectHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz proje ID", nil)
		return
	}

	var in services.ProjectInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz proje bilgileri", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.svc.Update(c.Request.Context(), id, in, userID)
	if err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Proje bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrCariNotFoundForProject) {
			utils.Err(c, http.StatusNotFound, "CARI_NOT_FOUND", "Cari kart bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz proje ID", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.svc.Delete(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Proje bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"deleted": true})
}

func (h *ProjectHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz proje ID", nil)
		return
	}

	res, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Proje bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *ProjectHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	query := c.Query("q")
	sort := c.Query("sort")

	filters := make(map[string]string)
	if statusFilter := c.Query("status"); statusFilter != "" {
		filters["status"] = statusFilter
	}
	if cariFilter := c.Query("cari_id"); cariFilter != "" {
		filters["cari_id"] = cariFilter
	}

	if limit > 100 {
		limit = 100
	}
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}

	res, total, err := h.svc.List(c.Request.Context(), page, limit, query, sort, filters)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.List(c, res, page, limit, total)
}

type addInvoiceInput struct {
	InvoiceID uuid.UUID `json:"invoice_id" binding:"required"`
}

func (h *ProjectHandler) AddInvoice(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz proje ID", nil)
		return
	}

	var in addInvoiceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz fatura bilgileri", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.svc.AddInvoiceToProject(c.Request.Context(), projectID, in.InvoiceID, userID)
	if err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Proje bulunamadı", nil)
			return
		}
		if err.Error() == "invoice_not_found" {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Fatura bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"added": true})
}

type addQuoteInput struct {
	QuoteID uuid.UUID `json:"quote_id" binding:"required"`
}

func (h *ProjectHandler) AddQuote(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz proje ID", nil)
		return
	}

	var in addQuoteInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz teklif bilgileri", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.svc.AddQuoteToProject(c.Request.Context(), projectID, in.QuoteID, userID)
	if err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Proje bulunamadı", nil)
			return
		}
		if err.Error() == "quote_not_found" {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Teklif bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"added": true})
}

type removeInvoiceInput struct {
	InvoiceID uuid.UUID `json:"invoice_id" binding:"required"`
}

func (h *ProjectHandler) RemoveInvoice(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz proje ID", nil)
		return
	}

	var in removeInvoiceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz fatura bilgileri", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.svc.RemoveInvoiceFromProject(c.Request.Context(), projectID, in.InvoiceID, userID)
	if err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Proje bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"removed": true})
}

type removeQuoteInput struct {
	QuoteID uuid.UUID `json:"quote_id" binding:"required"`
}

func (h *ProjectHandler) RemoveQuote(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz proje ID", nil)
		return
	}

	var in removeQuoteInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz teklif bilgileri", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.svc.RemoveQuoteFromProject(c.Request.Context(), projectID, in.QuoteID, userID)
	if err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Proje bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"removed": true})
}

type addEmployeeInput struct {
	EmployeeID uuid.UUID `json:"employee_id" binding:"required"`
}

func (h *ProjectHandler) AddEmployee(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz proje ID", nil)
		return
	}

	var in addEmployeeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz personel bilgileri", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.svc.AddEmployeeToProject(c.Request.Context(), projectID, in.EmployeeID, userID)
	if err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Proje bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"added": true})
}

type removeEmployeeInput struct {
	EmployeeID uuid.UUID `json:"employee_id" binding:"required"`
}

func (h *ProjectHandler) RemoveEmployee(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz proje ID", nil)
		return
	}

	var in removeEmployeeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz personel bilgileri", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.svc.RemoveEmployeeFromProject(c.Request.Context(), projectID, in.EmployeeID, userID)
	if err != nil {
		if errors.Is(err, services.ErrProjectNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Proje bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"removed": true})
}
