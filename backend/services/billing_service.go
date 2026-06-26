package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"radikal-hesap/database"
	"radikal-hesap/models"
)

type BillingService struct {
	provider PaymentProvider
}

func NewBillingService(provider PaymentProvider) *BillingService {
	return &BillingService{provider: provider}
}

const renewWindowDays = 15

type BillingStatusResponse struct {
	SubscriptionStatus string     `json:"subscription_status"`
	TrialEndsAt        time.Time  `json:"trial_ends_at"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end"`
	PlanName           string     `json:"plan_name"`
	PlanCode           string     `json:"plan_code"`
	TrialDaysLeft      int        `json:"trial_days_left"`
	PeriodDaysLeft     int        `json:"period_days_left"`
	CanRenew           bool       `json:"can_renew"`
}

type SubscribeInput struct {
	PlanID     uuid.UUID `json:"plan_id" binding:"required"`
	PeriodType string    `json:"period_type" binding:"required,oneof=monthly yearly"`
}

type SubscribeResponse struct {
	SessionID   string `json:"session_id"`
	CheckoutURL string `json:"checkout_url"`
}

func (s *BillingService) GetBillingStatus(ctx context.Context) (*BillingStatusResponse, error) {
	compIDStr, ok := ctx.Value("company_id").(string)
	if !ok || compIDStr == "" {
		// Handle cases where company_id is not set (e.g., superadmin)
		return &BillingStatusResponse{
			SubscriptionStatus: "active",
			PlanName:           "Superadmin",
			PlanCode:           "superadmin",
			TrialDaysLeft:      999,
		}, nil
	}
	companyID, err := uuid.Parse(compIDStr)
	if err != nil || companyID == uuid.Nil {
		return &BillingStatusResponse{
			SubscriptionStatus: "active",
			PlanName:           "Superadmin",
			PlanCode:           "superadmin",
			TrialDaysLeft:      999,
		}, nil
	}

	var company models.Company
	if err := database.SystemDB.First(&company, companyID).Error; err != nil {
		return nil, err
	}

	planName := "Trial Plan"
	planCode := "free"

	if company.PlanID != nil {
		var plan models.Plan
		if err := database.SystemDB.First(&plan, *company.PlanID).Error; err == nil {
			planName = plan.Name
			planCode = plan.Code
		}
	}

	daysLeft := 0
	if company.SubscriptionStatus == "trial" {
		diff := time.Until(company.TrialEndsAt)
		daysLeft = int(diff.Hours() / 24)
		if daysLeft < 0 {
			daysLeft = 0
		}
	}

	periodDaysLeft := 0
	canRenew := false
	if company.SubscriptionStatus == "active" && company.CurrentPeriodEnd != nil {
		diff := time.Until(*company.CurrentPeriodEnd)
		periodDaysLeft = int(diff.Hours() / 24)
		canRenew = periodDaysLeft <= renewWindowDays
	}

	return &BillingStatusResponse{
		SubscriptionStatus: company.SubscriptionStatus,
		TrialEndsAt:        company.TrialEndsAt,
		CurrentPeriodEnd:   company.CurrentPeriodEnd,
		PlanName:           planName,
		PlanCode:           planCode,
		TrialDaysLeft:      daysLeft,
		PeriodDaysLeft:     periodDaysLeft,
		CanRenew:           canRenew,
	}, nil
}

func (s *BillingService) GetPlans() ([]models.Plan, error) {
	var plans []models.Plan
	// Ücretsiz planlar (fiyatsız) satın alma ekranında gösterilmez; deneme süresi
	// kayıt anında otomatik tanımlanır, kullanıcı tekrar ücretsiz plan seçemez.
	if err := database.SystemDB.
		Where("is_active = ?", true).
		Where("price_monthly > 0 OR price_yearly > 0").
		Order("price_monthly asc").
		Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

func (s *BillingService) Subscribe(ctx context.Context, planID uuid.UUID, periodType string) (*SubscribeResponse, error) {
	compIDStr := ctx.Value("company_id").(string)
	companyID, err := uuid.Parse(compIDStr)
	if err != nil {
		return nil, err
	}

	// Validate Plan exists
	var plan models.Plan
	if err := database.SystemDB.First(&plan, planID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("plan_not_found")
		}
		return nil, err
	}

	var company models.Company
	if err := database.SystemDB.First(&company, companyID).Error; err != nil {
		return nil, err
	}

	// Aktif ücretli abonelik bitişine 15 günden fazla varken sadece daha yüksek
	// fiyatlı bir plana (upgrade) geçilebilir; aynı/daha düşük plana geçiş ve
	// yeniden abone olma engellenir. 15 gün penceresine girilince (CanRenew)
	// herhangi bir plana serbestçe geçilebilir.
	if company.SubscriptionStatus == "active" && company.PlanID != nil && company.CurrentPeriodEnd != nil {
		daysLeft := time.Until(*company.CurrentPeriodEnd).Hours() / 24
		if daysLeft > renewWindowDays {
			var currentPlan models.Plan
			if err := database.SystemDB.First(&currentPlan, *company.PlanID).Error; err == nil {
				// Upgrade kontrolü seçilen dönem tipinin fiyatı üzerinden yapılır;
				// aylık abonede aylık, yıllık abonede yıllık fiyat karşılaştırılır.
				newPrice := plan.PriceMonthly
				curPrice := currentPlan.PriceMonthly
				if periodType == "yearly" {
					newPrice = plan.PriceYearly
					curPrice = currentPlan.PriceYearly
				}
				if newPrice.LessThanOrEqual(curPrice) {
					return nil, errors.New("plan_change_not_allowed_yet")
				}
			}
		}
	}

	tx := database.SystemDB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var amount decimal.Decimal
	if periodType == "yearly" {
		amount = plan.PriceYearly
	} else {
		amount = plan.PriceMonthly
	}

	// Charge using payment provider
	chargeReq := ChargeRequest{
		CompanyID:  companyID,
		Amount:     amount,
		Currency:   plan.Currency,
		PlanID:     planID,
		PeriodType: periodType,
	}

	chargeRes, err := s.provider.Charge(ctx, chargeReq)
	txStatus := "success"
	var chargeErr error
	if err != nil || chargeRes == nil || !chargeRes.Success {
		txStatus = "failed"
		if err != nil {
			chargeErr = err
		} else if chargeRes != nil {
			chargeErr = errors.New(chargeRes.Message)
		} else {
			chargeErr = errors.New("payment_failed")
		}
	}

	periodStart := time.Now()
	var periodEnd time.Time
	if periodType == "yearly" {
		periodEnd = periodStart.AddDate(1, 0, 0)
	} else {
		periodEnd = periodStart.AddDate(0, 1, 0)
	}

	providerRef := ""
	if chargeRes != nil {
		providerRef = chargeRes.ProviderRef
	}

	billingTx := models.BillingTransaction{
		ID:          uuid.New(),
		CompanyID:   companyID,
		PlanID:      &planID,
		Action:      "subscribe",
		Amount:      amount,
		Currency:    plan.Currency,
		Status:      txStatus,
		ProviderRef: providerRef,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		CreatedAt:   time.Now(),
	}

	if err := tx.Create(&billingTx).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if txStatus == "failed" {
		tx.Commit()
		return nil, chargeErr
	}

	company.PlanID = &planID
	company.SubscriptionStatus = "active"
	company.CurrentPeriodEnd = &periodEnd

	if err := tx.Save(&company).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &SubscribeResponse{
		SessionID:   "direct_activation",
		CheckoutURL: "/billing?success=direct",
	}, nil
}

func (s *BillingService) Renew(ctx context.Context, periodType string) (*SubscribeResponse, error) {
	compIDStr := ctx.Value("company_id").(string)
	companyID, err := uuid.Parse(compIDStr)
	if err != nil {
		return nil, err
	}

	var company models.Company
	if err := database.SystemDB.First(&company, companyID).Error; err != nil {
		return nil, err
	}

	if company.PlanID == nil {
		return nil, errors.New("no_active_plan_to_renew")
	}

	if company.SubscriptionStatus == "active" && company.CurrentPeriodEnd != nil {
		daysLeft := time.Until(*company.CurrentPeriodEnd).Hours() / 24
		if daysLeft > renewWindowDays {
			return nil, errors.New("renew_not_yet_available")
		}
	}

	var plan models.Plan
	if err := database.SystemDB.First(&plan, *company.PlanID).Error; err != nil {
		return nil, err
	}

	var amount decimal.Decimal
	var periodEnd time.Time
	periodStart := time.Now()
	if company.CurrentPeriodEnd != nil && company.CurrentPeriodEnd.After(time.Now()) {
		periodStart = *company.CurrentPeriodEnd
	}

	if periodType == "yearly" {
		amount = plan.PriceYearly
		periodEnd = periodStart.AddDate(1, 0, 0)
	} else {
		amount = plan.PriceMonthly
		periodEnd = periodStart.AddDate(0, 1, 0)
	}

	tx := database.SystemDB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	chargeReq := ChargeRequest{
		CompanyID:  companyID,
		Amount:     amount,
		Currency:   plan.Currency,
		PlanID:     plan.ID,
		PeriodType: periodType,
	}

	chargeRes, err := s.provider.Charge(ctx, chargeReq)
	txStatus := "success"
	var chargeErr error
	if err != nil || chargeRes == nil || !chargeRes.Success {
		txStatus = "failed"
		if err != nil {
			chargeErr = err
		} else if chargeRes != nil {
			chargeErr = errors.New(chargeRes.Message)
		} else {
			chargeErr = errors.New("payment_failed")
		}
	}

	providerRef := ""
	if chargeRes != nil {
		providerRef = chargeRes.ProviderRef
	}

	billingTx := models.BillingTransaction{
		ID:          uuid.New(),
		CompanyID:   companyID,
		PlanID:      company.PlanID,
		Action:      "renew",
		Amount:      amount,
		Currency:    plan.Currency,
		Status:      txStatus,
		ProviderRef: providerRef,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		CreatedAt:   time.Now(),
	}

	if err := tx.Create(&billingTx).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if txStatus == "failed" {
		tx.Commit()
		return nil, chargeErr
	}

	company.SubscriptionStatus = "active"
	company.CurrentPeriodEnd = &periodEnd

	if err := tx.Save(&company).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &SubscribeResponse{
		SessionID:   "direct_activation",
		CheckoutURL: "/billing?success=renew",
	}, nil
}

func (s *BillingService) GetTransactions(ctx context.Context) ([]models.BillingTransaction, error) {
	compIDStr := ctx.Value("company_id").(string)
	companyID, err := uuid.Parse(compIDStr)
	if err != nil {
		return nil, err
	}

	var transactions []models.BillingTransaction
	if err := database.SystemDB.Where("company_id = ?", companyID).Order("created_at desc").Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

var (
	ErrWebhookInvalidSignature = errors.New("webhook_invalid_signature")
	ErrWebhookMissingEventID   = errors.New("webhook_missing_event_id")
)

type StripeWebhookPayload struct {
	EventID    string `json:"event_id"`
	Type       string `json:"type"` // e.g. "checkout.session.completed", "customer.subscription.deleted"
	CompanyID  string `json:"company_id"`
	PlanID     string `json:"plan_id"`
	PeriodType string `json:"period_type"` // "monthly" / "yearly"
	SessionID  string `json:"session_id"`
}

// ProcessWebhook handles incoming webhook events.
// rawBody ve signature, sağlayıcı (Stripe/banka) imza doğrulaması için ham
// gövde ve "Stripe-Signature" header değeridir.
// NOTE: This webhook format is Stripe-specific; actual bank integration
// will require a different endpoint/payload format later.
func (s *BillingService) ProcessWebhook(rawBody []byte, signature string, payload StripeWebhookPayload) error {
	// 1. İmza doğrulaması — kimliği doğrulanmamış istekler reddedilir.
	// (Public endpoint olduğu için bu, yetkisiz ücretsiz abonelik/iptal
	// saldırılarına karşı tek savunmadır.)
	valid, err := s.provider.VerifyWebhook(rawBody, signature)
	if err != nil {
		return err
	}
	if !valid {
		return ErrWebhookInvalidSignature
	}

	// 2. Idempotency — event_id zorunlu; aynı event tekrar gelirse atlanır.
	if payload.EventID == "" {
		return ErrWebhookMissingEventID
	}

	compID, err := uuid.Parse(payload.CompanyID)
	if err != nil {
		return errors.New("invalid_company_id")
	}

	var company models.Company
	if err := database.SystemDB.First(&company, compID).Error; err != nil {
		return err
	}

	tx := database.SystemDB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Bu event daha önce işlendiyse (Stripe retry vb.) hiçbir şey yapma.
	var existing int64
	if err := tx.Model(&models.BillingTransaction{}).Where("event_id = ?", payload.EventID).Count(&existing).Error; err != nil {
		tx.Rollback()
		return err
	}
	if existing > 0 {
		tx.Rollback()
		return nil // zaten işlenmiş — idempotent başarı
	}

	if payload.Type == "checkout.session.completed" {
		planID, err := uuid.Parse(payload.PlanID)
		if err != nil {
			tx.Rollback()
			return errors.New("invalid_plan_id")
		}

		var plan models.Plan
		if err := tx.First(&plan, planID).Error; err != nil {
			tx.Rollback()
			return errors.New("plan_not_found")
		}

		// Süre period_type'a göre; yearly için +1 yıl, aksi halde +1 ay.
		periodStart := time.Now()
		if company.CurrentPeriodEnd != nil && company.CurrentPeriodEnd.After(periodStart) {
			periodStart = *company.CurrentPeriodEnd
		}
		var periodEnd time.Time
		var amount decimal.Decimal
		if payload.PeriodType == "yearly" {
			periodEnd = periodStart.AddDate(1, 0, 0)
			amount = plan.PriceYearly
		} else {
			periodEnd = periodStart.AddDate(0, 1, 0)
			amount = plan.PriceMonthly
		}

		company.PlanID = &planID
		company.SubscriptionStatus = "active"
		company.CurrentPeriodEnd = &periodEnd
		if company.StripeSubscriptionID == "" {
			company.StripeSubscriptionID = fmt.Sprintf("sub_%s", uuid.NewString()[:8])
		}
		if company.StripeCustomerID == "" {
			company.StripeCustomerID = fmt.Sprintf("cus_%s", uuid.NewString()[:8])
		}

		if err := tx.Save(&company).Error; err != nil {
			tx.Rollback()
			return err
		}

		billingTx := models.BillingTransaction{
			ID:          uuid.New(),
			CompanyID:   compID,
			PlanID:      &planID,
			Action:      "subscribe",
			Amount:      amount,
			Currency:    plan.Currency,
			Status:      "success",
			ProviderRef: payload.SessionID,
			EventID:     payload.EventID,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			CreatedAt:   time.Now(),
		}
		if err := tx.Create(&billingTx).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else if payload.Type == "customer.subscription.deleted" {
		company.SubscriptionStatus = "canceled"
		company.PlanID = nil
		company.CurrentPeriodEnd = nil

		if err := tx.Save(&company).Error; err != nil {
			tx.Rollback()
			return err
		}

		billingTx := models.BillingTransaction{
			ID:        uuid.New(),
			CompanyID: compID,
			Action:    "cancel",
			Status:    "success",
			Currency:  company.Currency,
			EventID:   payload.EventID,
			CreatedAt: time.Now(),
		}
		if err := tx.Create(&billingTx).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// Bilinmeyen event türü — idempotency kaydı yaz, durumu değiştirme.
		billingTx := models.BillingTransaction{
			ID:        uuid.New(),
			CompanyID: compID,
			Action:    "webhook_ignored",
			Status:    "success",
			Currency:  company.Currency,
			EventID:   payload.EventID,
			CreatedAt: time.Now(),
		}
		if err := tx.Create(&billingTx).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}
