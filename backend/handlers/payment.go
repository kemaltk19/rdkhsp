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

type PaymentHandler struct {
	svc     *services.PaymentService
	cashSvc *services.CashAccountService
	bankSvc *services.BankAccountService
}

func NewPaymentHandler(svc *services.PaymentService, cashSvc *services.CashAccountService, bankSvc *services.BankAccountService) *PaymentHandler {
	return &PaymentHandler{
		svc:     svc,
		cashSvc: cashSvc,
		bankSvc: bankSvc,
	}
}

func (h *PaymentHandler) Create(c *gin.Context) {
	var in services.PaymentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz işlem verileri", nil)
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
		if errors.Is(err, services.ErrCariNotFoundForPayment) {
			utils.Err(c, http.StatusNotFound, "CARI_NOT_FOUND", "Cari kart bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrAccountNotFound) {
			utils.Err(c, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "Kasa veya Banka hesabı bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrCurrencyMismatch) {
			utils.Err(c, http.StatusConflict, "CURRENCY_MISMATCH", "Para birimleri uyuşmuyor", nil)
			return
		}
		if errors.Is(err, services.ErrInvoiceNotFoundForPay) {
			utils.Err(c, http.StatusNotFound, "INVOICE_NOT_FOUND", "Fatura bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.Created(c, res)
}

func (h *PaymentHandler) Cancel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz işlem ID", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.svc.Cancel(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, services.ErrPaymentNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "İşlem kaydı bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrPaymentAlreadyCanceled) {
			utils.Err(c, http.StatusConflict, "ALREADY_CANCELED", "Bu işlem zaten iptal edilmiş", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"canceled": true})
}

func (h *PaymentHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz işlem ID", nil)
		return
	}

	var in services.PaymentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz işlem verileri", nil)
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
		if errors.Is(err, services.ErrPaymentNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "İşlem kaydı bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrCariNotFoundForPayment) {
			utils.Err(c, http.StatusNotFound, "CARI_NOT_FOUND", "Cari kart bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrAccountNotFound) {
			utils.Err(c, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "Kasa veya Banka hesabı bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrCurrencyMismatch) {
			utils.Err(c, http.StatusConflict, "CURRENCY_MISMATCH", "Para birimleri uyuşmuyor", nil)
			return
		}
		if errors.Is(err, services.ErrInvoiceNotFoundForPay) {
			utils.Err(c, http.StatusNotFound, "INVOICE_NOT_FOUND", "Fatura bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *PaymentHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz işlem ID", nil)
		return
	}

	res, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrPaymentNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "İşlem bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *PaymentHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	query := c.Query("q")
	sort := c.Query("sort")

	filters := make(map[string]string)
	if typeFilter := c.Query("type"); typeFilter != "" {
		filters["type"] = typeFilter
	}
	if statusFilter := c.Query("status"); statusFilter != "" {
		filters["status"] = statusFilter
	}
	if cariFilter := c.Query("cari_id"); cariFilter != "" {
		filters["cari_id"] = cariFilter
	}
	if invoiceFilter := c.Query("invoice_id"); invoiceFilter != "" {
		filters["invoice_id"] = invoiceFilter
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

// ----------------------------------------------------
// Cash Account and Bank Account Actions
// ----------------------------------------------------

func (h *PaymentHandler) CreateCashAccount(c *gin.Context) {
	var in struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code"`
		AccountNo   string `json:"account_no"`
		Description string `json:"description"`
		Currency    string `json:"currency" binding:"required"`
		IsDefault   bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz kasa bilgileri", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.cashSvc.Create(c.Request.Context(), in.Name, in.Code, in.AccountNo, in.Description, in.Currency, in.IsDefault, userID)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.Created(c, res)
}

func (h *PaymentHandler) ListCashAccounts(c *gin.Context) {
	res, err := h.cashSvc.List(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	utils.OK(c, res)
}

func (h *PaymentHandler) CreateBankAccount(c *gin.Context) {
	var in struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code"`
		AccountNo   string `json:"account_no"`
		Description string `json:"description"`
		IBAN        string `json:"iban"`
		Currency    string `json:"currency" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz banka bilgileri", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.bankSvc.Create(c.Request.Context(), in.Name, in.Code, in.AccountNo, in.Description, in.IBAN, in.Currency, userID)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.Created(c, res)
}

func (h *PaymentHandler) ListBankAccounts(c *gin.Context) {
	res, err := h.bankSvc.List(c.Request.Context())
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	utils.OK(c, res)
}
