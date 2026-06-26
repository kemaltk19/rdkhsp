package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/services"
	"radikal-hesap/utils"
)

type PublicInvoiceHandler struct{}

func NewPublicInvoiceHandler() *PublicInvoiceHandler {
	return &PublicInvoiceHandler{}
}

func (h *PublicInvoiceHandler) GetByToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Token eksik", nil)
		return
	}

	var invoice models.Invoice
	// Public token erişimi tenant context'i taşımaz (ziyaretçi login değildir),
	// bu yüzden RLS uygulanan database.DB'de satır görünmez; RLS bypass eden
	// SystemDB kullanılır. Token zaten gizli/rastgele olduğu için güvenlidir.
	if err := database.SystemDB.Preload("Items").First(&invoice, "public_token = ?", token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Fatura bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	var cari models.Cari
	if err := database.SystemDB.First(&cari, "id = ?", invoice.CariID).Error; err == nil {
		// Embed basic cari details
	}

	// Sadece public için güvenli veriyi dön
	res := gin.H{
		"id":             invoice.ID,
		"number":         invoice.Number,
		"date":           invoice.Date,
		"due_date":       invoice.DueDate,
		"currency":       invoice.Currency,
		"total":          invoice.Total,
		"subtotal":       invoice.Subtotal,
		"tax_total":      invoice.TaxTotal,
		"discount_total": invoice.DiscountTotal,
		"paid_total":     invoice.PaidTotal,
		"status":         invoice.Status,
		"note":           invoice.Note,
		"dispute_note":   invoice.DisputeNote,
		"items":          invoice.Items,
		"cari": gin.H{
			"name":  cari.Name,
			"email": cari.Email,
			"phone": cari.Phone,
		},
	}

	utils.OK(c, res)
}

type DisputeInput struct {
	Note string `json:"note" binding:"required,max=2000"`
}

func (h *PublicInvoiceHandler) Dispute(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Token eksik", nil)
		return
	}

	var in DisputeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.Err(c, http.StatusUnprocessableEntity, "VALIDATION", "İtiraz notu zorunludur", nil)
		return
	}

	// Public token erişiminde tenant context'i (app.company_id) yoktur; RLS uygulanan
	// database.DB bu yüzden satırı gizler. Token zaten gizli/tekil olduğundan
	// SystemDB (RLS bypass) üzerinden işlem yapmak güvenlidir.
	err := database.SystemDB.Transaction(func(tx *gorm.DB) error {
		var invoice models.Invoice
		if err := tx.First(&invoice, "public_token = ?", token).Error; err != nil {
			return err
		}

		if invoice.Status == "paid" || invoice.Status == "canceled" || invoice.Status == "draft" {
			return errors.New("fatura bu statüde iken itiraz edilemez")
		}

		now := time.Now()
		if err := tx.Model(&invoice).Updates(map[string]interface{}{
			"status":       "disputed",
			"dispute_note": in.Note,
			"disputed_at":  now,
		}).Error; err != nil {
			return err
		}

		// Kalıcı bildirim oluştur
		notifSvc := services.NewNotificationService()
		title := "Fatura İtirazı: " + invoice.Number
		message := "Müşteri faturaya itiraz etti. Not: " + in.Note
		ctxWithTx := context.WithValue(c.Request.Context(), utils.TxKey, tx)
		if err := notifSvc.CreateNotification(ctxWithTx, tx, invoice.CompanyID, "invoice_dispute", title, message, &invoice.ID, "invoice"); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Fatura bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"disputed": true})
}

func (h *PublicInvoiceHandler) Pay(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		utils.Err(c, http.StatusBadRequest, "BAD_REQUEST", "Token eksik", nil)
		return
	}

	var paymentRes *models.Payment

	// Public token erişiminde tenant context'i (app.company_id) yoktur; RLS uygulanan
	// database.DB bu yüzden satırı gizler. Token zaten gizli/tekil olduğundan
	// SystemDB (RLS bypass) üzerinden işlem yapmak güvenlidir.
	err := database.SystemDB.Transaction(func(tx *gorm.DB) error {
		var invoice models.Invoice
		// 1. Fetch invoice with Items and lock row to prevent double payment / race conditions
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(&invoice, "public_token = ?", token).Error; err != nil {
			return err
		}

		if invoice.Status != "sent" && invoice.Status != "partial" {
			return errors.New("fatura bu statüde iken ödeme yapılamaz")
		}

		remainingAmount := invoice.Total.Sub(invoice.PaidTotal)
		if remainingAmount.LessThanOrEqual(decimal.Zero) {
			return errors.New("faturanın ödenecek tutarı kalmamıştır")
		}

		var accountID uuid.UUID
		var accountKind string

		// Try to find a bank account for company matching invoice currency
		var bankAcc models.BankAccount
		if err := tx.First(&bankAcc, "company_id = ? AND currency = ?", invoice.CompanyID, invoice.Currency).Error; err == nil {
			accountKind = "bank"
			accountID = bankAcc.ID
		} else {
			// Try to find a cash account
			var cashAcc models.CashAccount
			if err := tx.First(&cashAcc, "company_id = ? AND currency = ? AND is_default = true", invoice.CompanyID, invoice.Currency).Error; err == nil {
				accountKind = "cash"
				accountID = cashAcc.ID
			} else if err := tx.First(&cashAcc, "company_id = ? AND currency = ?", invoice.CompanyID, invoice.Currency).Error; err == nil {
				accountKind = "cash"
				accountID = cashAcc.ID
			} else {
				// Create a default Bank Account
				accID, _ := uuid.NewV7()
				bankAcc = models.BankAccount{
					ID:        accID,
					CompanyID: invoice.CompanyID,
					Code:      "ONL-PAY",
					Name:      "Online Ödemeler Hesabı",
					Currency:  invoice.Currency,
					Balance:   decimal.Zero,
				}
				if err := tx.Create(&bankAcc).Error; err != nil {
					return err
				}
				accountKind = "bank"
				accountID = bankAcc.ID
			}
		}

		ctx := context.WithValue(c.Request.Context(), "company_id", invoice.CompanyID.String())
		ctxWithTx := context.WithValue(ctx, utils.TxKey, tx)

		var creatorID uuid.UUID
		if invoice.CreatedBy != nil {
			creatorID = *invoice.CreatedBy
		} else {
			creatorID = uuid.Nil
		}

		paymentInput := services.PaymentInput{
			CariID:      invoice.CariID,
			Type:        "collection", // client pays us
			Date:        time.Now(),
			Method:      "card", // online credit card
			AccountKind: accountKind,
			AccountID:   accountID,
			Amount:      remainingAmount,
			Currency:    invoice.Currency,
			InvoiceID:   &invoice.ID,
			Reference:   "ONLINE-" + invoice.Number,
			Note:        "Müşteri tarafından public link üzerinden yapılan online ödeme.",
		}

		paymentSvc := services.NewPaymentService()
		var pErr error
		paymentRes, pErr = paymentSvc.Create(ctxWithTx, paymentInput, creatorID)
		if pErr != nil {
			return pErr
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Err(c, http.StatusNotFound, "NOT_FOUND", "Fatura bulunamadı", nil)
			return
		}
		utils.Err(c, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}

	utils.OK(c, gin.H{"paid": true, "payment": paymentRes})
}
