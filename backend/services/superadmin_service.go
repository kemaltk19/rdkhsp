package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
	"github.com/shopspring/decimal"
)

type SuperadminService struct{}

func NewSuperadminService() *SuperadminService {
	return &SuperadminService{}
}

// -----------------------------------------------------------------------------
// Company Management
// -----------------------------------------------------------------------------

type CompanyStats struct {
	TotalUsers    int64 `json:"total_users"`
	TotalInvoices int64 `json:"total_invoices"`
}

type CompanyWithStats struct {
	models.Company
	Stats      CompanyStats `json:"stats" gorm:"-"`
	AdminName  string       `json:"admin_name" gorm:"-"`
	AdminEmail string       `json:"admin_email" gorm:"-"`
}

type PlanDistributionItem struct {
	PlanName string `json:"plan_name"`
	Count    int64  `json:"count"`
}

type SuperadminDashboardStats struct {
	TotalCompanies     int64                  `json:"total_companies"`
	ActiveCompanies    int64                  `json:"active_companies"`
	TrialCompanies     int64                  `json:"trial_companies"`
	TotalUsers         int64                  `json:"total_users"`
	MRR                decimal.Decimal        `json:"mrr"`
	Churn              int64                  `json:"churn"`
	PlanDistribution   []PlanDistributionItem `json:"plan_distribution"`
	RecentCompanies    []CompanyWithStats     `json:"recent_companies"`
}

func (s *SuperadminService) GetDashboardStats(ctx context.Context) (*SuperadminDashboardStats, error) {
	var stats SuperadminDashboardStats

	if err := database.SystemDB.Model(&models.Company{}).Count(&stats.TotalCompanies).Error; err != nil {
		return nil, err
	}

	if err := database.SystemDB.Model(&models.Company{}).Where("subscription_status = ?", "active").Count(&stats.ActiveCompanies).Error; err != nil {
		return nil, err
	}

	if err := database.SystemDB.Model(&models.Company{}).Where("subscription_status = ?", "trial").Count(&stats.TrialCompanies).Error; err != nil {
		return nil, err
	}

	if err := database.SystemDB.Model(&models.User{}).Where("role != ?", "superadmin").Count(&stats.TotalUsers).Error; err != nil {
		return nil, err
	}

	// MRR calculation
	var activeCompanies []models.Company
	if err := database.SystemDB.Where("subscription_status = ?", "active").Find(&activeCompanies).Error; err == nil {
		totalMRR := decimal.Zero
		for _, c := range activeCompanies {
			if c.PlanID != nil {
				var plan models.Plan
				if err := database.SystemDB.First(&plan, "id = ?", *c.PlanID).Error; err == nil {
					totalMRR = totalMRR.Add(plan.PriceMonthly)
					if plan.PriceYearly.GreaterThan(decimal.Zero) {
						// If they had a yearly billing cycle we could check that, but the prompt says:
						// active companies' Plan.PriceMonthly total (yearly plan -> PriceYearly/12). Let's check:
						// Let's assume standard monthly pricing, but if PriceMonthly is 0 and PriceYearly > 0, we do PriceYearly/12.
						if plan.PriceMonthly.IsZero() {
							totalMRR = totalMRR.Add(plan.PriceYearly.Div(decimal.NewFromInt(12)))
						}
					}
				}
			}
		}
		stats.MRR = totalMRR
	}

	// Churn: companies that changed to canceled (suspended) in past 30 days
	var churnCount int64
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	database.SystemDB.Model(&models.AuditLog{}).
		Where("module = ? AND action = ? AND created_at >= ?", "company", "suspend", thirtyDaysAgo).
		Count(&churnCount)
	stats.Churn = churnCount

	// Plan distribution
	var distributions []PlanDistributionItem
	rows, err := database.SystemDB.Table("companies").
		Select("plans.name as plan_name, count(*) as count").
		Joins("left join plans on plans.id = companies.plan_id").
		Group("plans.name").
		Rows()
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item PlanDistributionItem
			if err := rows.Scan(&item.PlanName, &item.Count); err == nil {
				if item.PlanName == "" {
					item.PlanName = "Paketsiz / Deneme"
				}
				distributions = append(distributions, item)
			}
		}
	}
	stats.PlanDistribution = distributions

	var recent []models.Company
	if err := database.SystemDB.Order("created_at desc").Limit(5).Find(&recent).Error; err != nil {
		return nil, err
	}

	var recentWithStats []CompanyWithStats
	for _, comp := range recent {
		var userCount int64
		database.SystemDB.Model(&models.User{}).Where("company_id = ?", comp.ID).Count(&userCount)

		var invoiceCount int64
		database.SystemDB.Model(&models.Invoice{}).Where("company_id = ?", comp.ID).Count(&invoiceCount)

		var adminName, adminEmail string
		var adminUser models.User
		if err := database.SystemDB.Where("company_id = ? AND role = ?", comp.ID, "admin").First(&adminUser).Error; err == nil {
			adminName = adminUser.Name
			adminEmail = adminUser.Email
		}

		recentWithStats = append(recentWithStats, CompanyWithStats{
			Company: comp,
			Stats: CompanyStats{
				TotalUsers:    userCount,
				TotalInvoices: invoiceCount,
			},
			AdminName:  adminName,
			AdminEmail: adminEmail,
		})
	}
	stats.RecentCompanies = recentWithStats

	return &stats, nil
}

func (s *SuperadminService) GetCompanies(ctx context.Context) ([]CompanyWithStats, error) {
	var companies []models.Company
	if err := database.SystemDB.Order("created_at desc").Find(&companies).Error; err != nil {
		return nil, err
	}

	var result []CompanyWithStats
	for _, comp := range companies {
		var userCount int64
		database.SystemDB.Model(&models.User{}).Where("company_id = ?", comp.ID).Count(&userCount)

		var invoiceCount int64
		database.SystemDB.Model(&models.Invoice{}).Where("company_id = ?", comp.ID).Count(&invoiceCount)

		var adminName, adminEmail string
		var adminUser models.User
		if err := database.SystemDB.Where("company_id = ? AND role = ?", comp.ID, "admin").First(&adminUser).Error; err == nil {
			adminName = adminUser.Name
			adminEmail = adminUser.Email
		}

		result = append(result, CompanyWithStats{
			Company: comp,
			Stats: CompanyStats{
				TotalUsers:    userCount,
				TotalInvoices: invoiceCount,
			},
			AdminName:  adminName,
			AdminEmail: adminEmail,
		})
	}

	return result, nil
}

func (s *SuperadminService) ToggleCompanyStatus(ctx context.Context, id uuid.UUID, action string) error {
	var company models.Company
	if err := database.SystemDB.First(&company, id).Error; err != nil {
		return errors.New("company not found")
	}

	if action == "suspend" {
		company.SubscriptionStatus = "canceled"
	} else if action == "activate" {
		// Just a simple activate, but usually depends on their trial or sub status. Let's say trial or active.
		company.SubscriptionStatus = "active"
		// current_period_end boşsa (örn. hiç fatura akışından geçmemiş şirket) NULL
		// kalırsa bitiş tarihi hesaplamaları 0 günde kalıp "süresi doldu" gösterir.
		if company.CurrentPeriodEnd == nil || company.CurrentPeriodEnd.Before(time.Now()) {
			periodEnd := time.Now().AddDate(0, 1, 0)
			company.CurrentPeriodEnd = &periodEnd
		}
	} else {
		return errors.New("invalid action")
	}

	if err := database.SystemDB.Save(&company).Error; err != nil {
		return err
	}

	// Write audit log
	superadminUserIDStr, _ := ctx.Value("user_id").(string)
	superadminUserID, _ := uuid.Parse(superadminUserIDStr)
	summary := fmt.Sprintf("Firma durumu değiştirildi: %s", action)
	_ = WriteAuditLog(ctx, database.SystemDB, "company", id, action, superadminUserID, summary)

	return nil
}

// -----------------------------------------------------------------------------
// Plan Management
// -----------------------------------------------------------------------------

func (s *SuperadminService) GetPlans(ctx context.Context) ([]models.Plan, error) {
	var plans []models.Plan
	if err := database.SystemDB.Order("price_monthly asc").Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

func (s *SuperadminService) CreatePlan(ctx context.Context, plan *models.Plan) error {
	plan.ID = uuid.New() // Ensure UUID v7 logic if possible, or simple New
	plan.CreatedAt = time.Now()
	plan.UpdatedAt = time.Now()

	if plan.Features == "" {
		plan.Features = "{}"
	}

	return database.SystemDB.Create(plan).Error
}

func (s *SuperadminService) UpdatePlan(ctx context.Context, id uuid.UUID, data map[string]interface{}) error {
	data["updated_at"] = time.Now()
	return database.SystemDB.Model(&models.Plan{}).Where("id = ?", id).Updates(data).Error
}

func (s *SuperadminService) DeletePlan(ctx context.Context, id uuid.UUID) error {
	var count int64
	if err := database.SystemDB.Model(&models.Company{}).Where("plan_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return database.SystemDB.Model(&models.Plan{}).Where("id = ?", id).Update("is_active", false).Error
	}
	return database.SystemDB.Delete(&models.Plan{}, id).Error
}

// -----------------------------------------------------------------------------
// Email (SMTP) Settings — single platform-wide row, superadmin-managed
// -----------------------------------------------------------------------------

type EmailSettingsInput struct {
	Host      string `json:"host" binding:"max=255"`
	Port      string `json:"port" binding:"max=10"`
	Username  string `json:"username" binding:"max=255"`
	Password  string `json:"password" binding:"max=255"` // blank = keep existing
	FromEmail string `json:"from_email" binding:"max=255"`
	FromName  string `json:"from_name" binding:"max=255"`
	Enabled   bool   `json:"enabled"`
}

func (s *SuperadminService) GetEmailSettings(ctx context.Context) (*models.EmailSetting, error) {
	var es models.EmailSetting
	err := database.SystemDB.First(&es).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.EmailSetting{}, nil
		}
		return nil, err
	}
	return &es, nil
}

func (s *SuperadminService) SaveEmailSettings(ctx context.Context, in EmailSettingsInput) (*models.EmailSetting, error) {
	var es models.EmailSetting
	err := database.SystemDB.First(&es).Error
	isNew := false
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			isNew = true
			es.ID = uuid.New()
		} else {
			return nil, err
		}
	}

	es.Host = in.Host
	es.Port = in.Port
	es.Username = in.Username
	es.FromEmail = in.FromEmail
	es.FromName = in.FromName
	es.Enabled = in.Enabled
	es.UpdatedAt = time.Now()

	if in.Password != "" {
		enc, encErr := utils.Encrypt(in.Password)
		if encErr != nil {
			return nil, encErr
		}
		es.PasswordEnc = enc
	}

	if isNew {
		if err := database.SystemDB.Create(&es).Error; err != nil {
			return nil, err
		}
	} else {
		if err := database.SystemDB.Save(&es).Error; err != nil {
			return nil, err
		}
	}
	return &es, nil
}

// SendTestEmail sends a real SMTP test using the saved DB settings. Unlike the
// normal SendEmail (which silently falls back to the log mailer when SMTP is
// disabled/unconfigured), this must fail loudly: the whole point of the test
// button is to confirm the configured server actually accepts the mail.
func (s *SuperadminService) SendTestEmail(ctx context.Context, to string) error {
	var es models.EmailSetting
	if err := database.SystemDB.First(&es).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("e-posta ayarları henüz kaydedilmemiş")
		}
		return err
	}
	if !es.Enabled {
		return errors.New("e-posta gönderimi 'Aktif' değil — önce ayarları etkinleştirip kaydet")
	}
	if es.Host == "" {
		return errors.New("SMTP sunucu adresi boş")
	}

	pass, _ := utils.Decrypt(es.PasswordEnc)
	cfg := utils.SMTPConfig{
		Host:     es.Host,
		Port:     es.Port,
		User:     es.Username,
		Pass:     pass,
		From:     es.FromEmail,
		FromName: es.FromName,
	}

	e := utils.Email{
		To:      to,
		Subject: "Radikal Hesap — SMTP Test",
		HTML:    "<p>SMTP ayarların çalışıyor. Bu bir test e-postasıdır.</p><p style=\"color:#9ca3af;font-size:12px;\">Radikal Hesap</p>",
	}
	return utils.SendSMTP(cfg, e)
}

type CreateCompanyInput struct {
	Name               string     `json:"name" binding:"required,max=255"`
	Email              string     `json:"email" binding:"max=255"`
	Phone              string     `json:"phone" binding:"max=50"`
	TaxOffice          string     `json:"tax_office" binding:"max=255"`
	TaxNumber          string     `json:"tax_number" binding:"max=50"`
	Address            string     `json:"address"`
	Currency           string     `json:"currency" binding:"required,max=10"`
	Locale             string     `json:"locale" binding:"required,max=10"`
	SubscriptionStatus string     `json:"subscription_status" binding:"max=50"`
	PlanID             *uuid.UUID `json:"plan_id"`
	// Additional company details
	ContactName string `json:"contact_name" binding:"max=255"`
	Landline    string `json:"landline" binding:"max=50"`
	Fax         string `json:"fax" binding:"max=50"`
	Industry    string `json:"industry" binding:"max=255"`
	Country     string `json:"country" binding:"max=255"`
	City        string `json:"city" binding:"max=255"`
	District    string `json:"district" binding:"max=255"`
	Timezone    string `json:"timezone" binding:"max=64"`
	// Initial Admin User
	AdminName     string `json:"admin_name" binding:"required,max=255"`
	AdminEmail    string `json:"admin_email" binding:"required,email,max=255"`
	AdminPassword string `json:"admin_password" binding:"required,min=8"`
}

type SuperadminUpdateCompanyInput struct {
	Name               string     `json:"name" binding:"required,max=255"`
	Email              string     `json:"email" binding:"max=255"`
	Phone              string     `json:"phone" binding:"max=50"`
	TaxOffice          string     `json:"tax_office" binding:"max=255"`
	TaxNumber          string     `json:"tax_number" binding:"max=50"`
	Address            string     `json:"address"`
	Currency           string     `json:"currency" binding:"required,max=10"`
	Locale             string     `json:"locale" binding:"required,max=10"`
	SubscriptionStatus string     `json:"subscription_status" binding:"required,max=50"`
	PlanID             *uuid.UUID `json:"plan_id"`
	// Additional company details
	ContactName string `json:"contact_name" binding:"max=255"`
	Landline    string `json:"landline" binding:"max=50"`
	Fax         string `json:"fax" binding:"max=50"`
	Industry    string `json:"industry" binding:"max=255"`
	Country     string `json:"country" binding:"max=255"`
	City        string `json:"city" binding:"max=255"`
	District    string `json:"district" binding:"max=255"`
	Timezone    string `json:"timezone" binding:"max=64"`
}

func (s *SuperadminService) CreateCompany(ctx context.Context, in CreateCompanyInput) (*models.Company, error) {
	email := strings.ToLower(strings.TrimSpace(in.AdminEmail))

	var existingUser models.User
	err := database.SystemDB.Where("email = ?", email).First(&existingUser).Error
	if err == nil {
		return nil, errors.New("email_exists")
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

	companyID, _ := uuid.NewV7()
	company := models.Company{
		ID:                 companyID,
		Name:               in.Name,
		Slug:               uniqueSlug(tx, in.Name),
		Email:              in.Email,
		Phone:              in.Phone,
		TaxOffice:          in.TaxOffice,
		TaxNumber:          in.TaxNumber,
		Address:            in.Address,
		Currency:           in.Currency,
		Locale:             in.Locale,
		SubscriptionStatus: in.SubscriptionStatus,
		PlanID:             in.PlanID,
		TrialEndsAt:        time.Now().AddDate(0, 0, 30),
		ContactName:        in.ContactName,
		Landline:           in.Landline,
		Fax:                in.Fax,
		Industry:           in.Industry,
		Country:            in.Country,
		City:               in.City,
		District:           in.District,
		Timezone:           in.Timezone,
	}
	if company.Timezone == "" {
		company.Timezone = "Europe/Istanbul"
	}
	if company.SubscriptionStatus == "" {
		company.SubscriptionStatus = "trial"
	}

	if err := tx.Create(&company).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	userID, _ := uuid.NewV7()
	passHash, err := utils.HashPassword(in.AdminPassword)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	user := models.User{
		ID:           userID,
		CompanyID:    &company.ID,
		Name:         in.AdminName,
		Email:        email,
		PasswordHash: passHash,
		Role:         "admin",
		Locale:       in.Locale,
		IsActive:     true,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		if isUniqueViolation(err) {
			return nil, errors.New("email_exists")
		}
		return nil, err
	}

	sequences := []models.NumberSequence{
		{CompanyID: company.ID, Key: "invoice_sales", LastNo: 0},
		{CompanyID: company.ID, Key: "invoice_purchase", LastNo: 0},
		{CompanyID: company.ID, Key: "quote", LastNo: 0},
		{CompanyID: company.ID, Key: "cari", LastNo: 0},
	}
	if err := tx.Create(&sequences).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	categories := []models.ProductCategory{
		{CompanyID: company.ID, Name: "Ticari Mallar"},
		{CompanyID: company.ID, Name: "Hammaddeler"},
		{CompanyID: company.ID, Name: "Yarı Mamuller"},
		{CompanyID: company.ID, Name: "Mamuller"},
		{CompanyID: company.ID, Name: "Sarf Malzemeleri"},
		{CompanyID: company.ID, Name: "Hizmet Ürünleri"},
		{CompanyID: company.ID, Name: "Demirbaşlar"},
	}
	for i := range categories {
		categories[i].ID, _ = uuid.NewV7()
	}
	if err := tx.Create(&categories).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 1) Seed Default Settings (Cari Groups, Quote Prefix)
	setting1ID, _ := uuid.NewV7()
	setting2ID, _ := uuid.NewV7()
	settings := []models.Setting{
		{ID: setting1ID, CompanyID: company.ID, Key: "cari_groups", Value: `["Bireysel", "Kurumsal", "Kurum", "Fabrika", "Esnaf", "Şirket", "Diğer"]`, Category: "cari"},
		{ID: setting2ID, CompanyID: company.ID, Key: "quote_prefix", Value: "PRO", Category: "quote"},
	}
	if err := tx.Create(&settings).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 2) Seed Expense Categories
	expenseCats := []models.ExpenseCategory{
		{CompanyID: company.ID, Name: "Elektrik"},
		{CompanyID: company.ID, Name: "Su"},
		{CompanyID: company.ID, Name: "Doğalgaz"},
		{CompanyID: company.ID, Name: "İnternet"},
		{CompanyID: company.ID, Name: "Telefon"},
		{CompanyID: company.ID, Name: "Temizlik"},
		{CompanyID: company.ID, Name: "Kargo"},
		{CompanyID: company.ID, Name: "Yemek"},
	}
	for i := range expenseCats {
		expenseCats[i].ID, _ = uuid.NewV7()
	}
	if err := tx.Create(&expenseCats).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 3) Seed Cash Accounts
	cashAccounts := []models.CashAccount{
		{CompanyID: company.ID, Code: "KAS-TL", Name: "Merkez Kasa (TRY)", Currency: "TRY", Balance: decimal.Zero, IsDefault: true},
		{CompanyID: company.ID, Code: "KAS-USD", Name: "Merkez Kasa (USD)", Currency: "USD", Balance: decimal.Zero, IsDefault: false},
		{CompanyID: company.ID, Code: "KAS-EUR", Name: "Merkez Kasa (EUR)", Currency: "EUR", Balance: decimal.Zero, IsDefault: false},
		{CompanyID: company.ID, Code: "KAS-GBP", Name: "Merkez Kasa (GBP)", Currency: "GBP", Balance: decimal.Zero, IsDefault: false},
		{CompanyID: company.ID, Code: "KAS-RUB", Name: "Merkez Kasa (RUB)", Currency: "RUB", Balance: decimal.Zero, IsDefault: false},
	}
	for i := range cashAccounts {
		cashAccounts[i].ID, _ = uuid.NewV7()
	}
	if err := tx.Create(&cashAccounts).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 4) Seed Bank Accounts
	bankAccounts := []models.BankAccount{
		{CompanyID: company.ID, Code: "BNK-TL", Name: "Banka Hesabı (TRY)", Currency: "TRY", Balance: decimal.Zero},
		{CompanyID: company.ID, Code: "BNK-USD", Name: "Banka Hesabı (USD)", Currency: "USD", Balance: decimal.Zero},
		{CompanyID: company.ID, Code: "BNK-EUR", Name: "Banka Hesabı (EUR)", Currency: "EUR", Balance: decimal.Zero},
		{CompanyID: company.ID, Code: "BNK-GBP", Name: "Banka Hesabı (GBP)", Currency: "GBP", Balance: decimal.Zero},
		{CompanyID: company.ID, Code: "BNK-RUB", Name: "Banka Hesabı (RUB)", Currency: "RUB", Balance: decimal.Zero},
	}
	for i := range bankAccounts {
		bankAccounts[i].ID, _ = uuid.NewV7()
	}
	if err := tx.Create(&bankAccounts).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 5) Seed Currencies
	currencies := []models.Currency{
		{CompanyID: company.ID, Name: "Türk Lirası", Symbol: "₺", Code: "TRY", ExchangeRate: decimal.NewFromInt(1), FormatPosition: "RightSpace", FormatThousandSep: ".", FormatDecimalSep: ",", FormatDecimals: 2, IsDefault: true},
		{CompanyID: company.ID, Name: "Amerikan Doları", Symbol: "$", Code: "USD", ExchangeRate: decimal.NewFromInt(1), FormatPosition: "Left", FormatThousandSep: ",", FormatDecimalSep: ".", FormatDecimals: 2, IsDefault: false},
		{CompanyID: company.ID, Name: "Euro", Symbol: "€", Code: "EUR", ExchangeRate: decimal.NewFromInt(1), FormatPosition: "RightSpace", FormatThousandSep: ".", FormatDecimalSep: ",", FormatDecimals: 2, IsDefault: false},
		{CompanyID: company.ID, Name: "İngiliz Sterlini", Symbol: "£", Code: "GBP", ExchangeRate: decimal.NewFromInt(1), FormatPosition: "Left", FormatThousandSep: ",", FormatDecimalSep: ".", FormatDecimals: 2, IsDefault: false},
		{CompanyID: company.ID, Name: "Rus Rublesi", Symbol: "₽", Code: "RUB", ExchangeRate: decimal.NewFromInt(1), FormatPosition: "RightSpace", FormatThousandSep: ".", FormatDecimalSep: ",", FormatDecimals: 2, IsDefault: false},
	}
	for i := range currencies {
		currencies[i].ID, _ = uuid.NewV7()
	}
	if err := tx.Create(&currencies).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 6) Seed Default Roles
	modules := []string{"caris", "invoices", "payments", "expenses", "products", "reports"}
	rolesData := []struct {
		Name        string
		Description string
		CanCreate   bool
		CanRead     bool
		CanUpdate   bool
		CanDelete   bool
	}{
		{Name: "İzle", Description: "Tüm modülleri sadece görüntüleme yetkisi", CanCreate: false, CanRead: true, CanUpdate: false, CanDelete: false},
		{Name: "Oluştur-İzle-Düzenle", Description: "Tüm modülleri görüntüleme, ekleme ve düzenleme yetkisi", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: false},
		{Name: "Oluştur-İzle-Düzenle-Sil", Description: "Tüm modüllerde tam yetki (ekleme, okuma, düzenleme ve silme)", CanCreate: true, CanRead: true, CanUpdate: true, CanDelete: true},
	}

	for _, rd := range rolesData {
		roleID, _ := uuid.NewV7()
		newRole := models.Role{
			ID:          roleID,
			CompanyID:   company.ID,
			Name:        rd.Name,
			Description: rd.Description,
		}
		if err := tx.Create(&newRole).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		var perms []models.RolePermission
		for _, m := range modules {
			permID, _ := uuid.NewV7()
			perms = append(perms, models.RolePermission{
				ID:        permID,
				RoleID:    roleID,
				Module:    m,
				CanCreate: rd.CanCreate,
				CanRead:   rd.CanRead,
				CanUpdate: rd.CanUpdate,
				CanDelete: rd.CanDelete,
			})
		}
		if err := tx.Create(&perms).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &company, nil
}

func (s *SuperadminService) UpdateCompany(ctx context.Context, id uuid.UUID, in SuperadminUpdateCompanyInput) (*models.Company, error) {
	var company models.Company
	if err := database.SystemDB.First(&company, id).Error; err != nil {
		return nil, errors.New("company not found")
	}

	company.Name = in.Name
	company.Email = in.Email
	company.Phone = in.Phone
	company.TaxOffice = in.TaxOffice
	company.TaxNumber = in.TaxNumber
	company.Address = in.Address
	company.Currency = in.Currency
	company.Locale = in.Locale
	company.SubscriptionStatus = in.SubscriptionStatus
	company.PlanID = in.PlanID
	company.ContactName = in.ContactName
	company.Landline = in.Landline
	company.Fax = in.Fax
	company.Industry = in.Industry
	company.Country = in.Country
	company.City = in.City
	company.District = in.District
	company.Timezone = in.Timezone
	if company.Timezone == "" {
		company.Timezone = "Europe/Istanbul"
	}

	// current_period_end boşsa (örn. hiç fatura akışından geçmemiş şirket) NULL
	// kalırsa bitiş tarihi hesaplamaları 0 günde kalıp "süresi doldu" gösterir.
	if in.SubscriptionStatus == "active" && (company.CurrentPeriodEnd == nil || company.CurrentPeriodEnd.Before(time.Now())) {
		periodEnd := time.Now().AddDate(0, 1, 0)
		company.CurrentPeriodEnd = &periodEnd
	} else if in.SubscriptionStatus != "active" {
		company.CurrentPeriodEnd = nil
	}
	company.UpdatedAt = time.Now()

	if err := database.SystemDB.Save(&company).Error; err != nil {
		return nil, err
	}

	// Write audit log
	superadminUserIDStr, _ := ctx.Value("user_id").(string)
	superadminUserID, _ := uuid.Parse(superadminUserIDStr)
	summary := fmt.Sprintf("Firma bilgileri güncellendi: %s", company.Name)
	_ = WriteAuditLog(ctx, database.SystemDB, "company", id, "update", superadminUserID, summary)

	return &company, nil
}

func (s *SuperadminService) DeleteCompany(ctx context.Context, id uuid.UUID) error {
	var company models.Company
	_ = database.SystemDB.First(&company, id)

	tx := database.SystemDB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// First, delete cari_balances since they reference cari_id and do not have company_id directly.
	if err := tx.Exec("DELETE FROM cari_balances WHERE cari_id IN (SELECT id FROM caris WHERE company_id = ?)", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	tables := []string{
		"invoice_items", "quote_items", "stock_movements", "cari_transactions", "cash_transactions",
		"invoices", "quotes", "payments", "expenses",
		"caris", "products", "employees", "warehouses", "cash_accounts", "bank_accounts",
		"product_categories", "expense_categories", "settings", "number_sequences",
	}

	for _, table := range tables {
		if err := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE company_id = ?", table), id).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Exec("DELETE FROM refresh_tokens WHERE user_id IN (SELECT id FROM users WHERE company_id = ?)", id).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Exec("DELETE FROM verification_tokens WHERE user_id IN (SELECT id FROM users WHERE company_id = ?)", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Exec("DELETE FROM users WHERE company_id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&models.Company{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Write audit log
	superadminUserIDStr, _ := ctx.Value("user_id").(string)
	superadminUserID, _ := uuid.Parse(superadminUserIDStr)
	summary := fmt.Sprintf("Firma tamamen silindi: %s", company.Name)
	_ = WriteAuditLog(ctx, database.SystemDB, "company", id, "delete", superadminUserID, summary)

	return nil
}
