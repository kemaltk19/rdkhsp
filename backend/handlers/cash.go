package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"radikal-hesap/services"
	"radikal-hesap/utils"
)

type CashHandler struct {
	svc *services.CashService
}

func NewCashHandler(svc *services.CashService) *CashHandler {
	return &CashHandler{svc: svc}
}

func (h *CashHandler) UpdateCashAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz kasa ID", nil)
		return
	}

	var in struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code"`
		AccountNo   string `json:"account_no"`
		Description string `json:"description"`
		Currency    string `json:"currency" binding:"required"`
		IsDefault   bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz veri", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.svc.UpdateCashAccount(c.Request.Context(), id, in.Name, in.Code, in.AccountNo, in.Description, in.Currency, in.IsDefault, userID)
	if err != nil {
		if errors.Is(err, services.ErrCashAccountNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Kasa hesabı bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *CashHandler) DeleteCashAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz kasa ID", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.svc.DeleteCashAccount(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, services.ErrCashAccountNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Kasa hesabı bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrAccountInUse) {
			utils.Err(c, http.StatusConflict, "CONFLICT", "Hareket görmüş kasa hesabı silinemez", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"deleted": true})
}

func (h *CashHandler) UpdateBankAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz banka ID", nil)
		return
	}

	var in struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code"`
		AccountNo   string `json:"account_no"`
		Description string `json:"description"`
		IBAN        string `json:"iban"`
		Currency    string `json:"currency" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz veri", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.svc.UpdateBankAccount(c.Request.Context(), id, in.Name, in.Code, in.AccountNo, in.Description, in.IBAN, in.Currency, userID)
	if err != nil {
		if errors.Is(err, services.ErrBankAccountNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Banka hesabı bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, res)
}

func (h *CashHandler) DeleteBankAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz banka ID", nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	err = h.svc.DeleteBankAccount(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, services.ErrBankAccountNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Banka hesabı bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrAccountInUse) {
			utils.Err(c, http.StatusConflict, "CONFLICT", "Hareket görmüş banka hesabı silinemez", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"deleted": true})
}

func (h *CashHandler) Transfer(c *gin.Context) {
	var in services.TransferInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "Geçersiz veri: "+err.Error(), nil)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		utils.Err(c, http.StatusUnauthorized, "UNAUTHORIZED", "Oturum açmanız gerekiyor", nil)
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	res, err := h.svc.Transfer(c.Request.Context(), in, userID)
	if err != nil {
		if errors.Is(err, services.ErrCashAccountNotFound) || errors.Is(err, services.ErrBankAccountNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Hesap bulunamadı", nil)
			return
		}
		if errors.Is(err, services.ErrTransferSameAccount) {
			utils.Err(c, http.StatusConflict, "SAME_ACCOUNT", "Aynı hesaba transfer yapılamez", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.Created(c, res)
}

func (h *CashHandler) ListTransactions(c *gin.Context) {
	kind := c.Query("account_kind")
	idStr := c.Query("account_id")

	if kind == "" || idStr == "" {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "account_kind ve account_id parametreleri zorunludur", nil)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz hesap ID", nil)
		return
	}

	res, err := h.svc.ListTransactions(c.Request.Context(), kind, id)
	if err != nil {
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.List(c, res, 1, len(res), int64(len(res)))
}
