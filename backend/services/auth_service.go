package services

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)

var (
	ErrEmailExists  = errors.New("email_exists")
	ErrInvalidCreds = errors.New("invalid_credentials")
	ErrUserInactive = errors.New("user_inactive")
	ErrInvalidToken = errors.New("invalid_token")
	ErrTokenExpired = errors.New("token_expired")
	ErrTokenRevoked = errors.New("token_revoked")
	ErrUserNotFound = errors.New("user_not_found")
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

type RegisterInput struct {
	CompanyName string `json:"company_name" binding:"required,max=255"`
	Name        string `json:"name" binding:"required,max=255"`
	Email       string `json:"email" binding:"required,email,max=255"`
	Password    string `json:"password" binding:"required,min=8"`
	Phone       string `json:"phone" binding:"max=50"`
	Industry    string `json:"industry" binding:"max=255"`
	Country     string `json:"country" binding:"max=255"`
	City        string `json:"city" binding:"max=255"`
	District    string `json:"district" binding:"max=255"`
	Address     string `json:"address"`
	Currency    string `json:"currency" binding:"max=10"`
	Locale      string `json:"locale" binding:"max=10"`
	Timezone    string `json:"timezone" binding:"max=64"`
}

type AuthResult struct {
	User         models.User             `json:"user"`
	Company      models.Company          `json:"company"`
	Permissions  []models.RolePermission `json:"permissions,omitempty"`
	AccessToken  string                  `json:"access_token"`
	RefreshToken string                  `json:"refresh_token"`
}

// loadPermissions, kullanıcının RoleID'si varsa o role ait modül izinlerini döner.
// admin/superadmin/cari veya RoleID=nil personel için boş liste döner (deny-by-default).
func loadPermissions(user models.User) []models.RolePermission {
	if user.RoleID == nil {
		return nil
	}
	var perms []models.RolePermission
	database.SystemDB.Where("role_id = ?", *user.RoleID).Find(&perms)
	return perms
}

func (s *AuthService) Register(in RegisterInput) (*AuthResult, error) {
	// Clean email input
	email := strings.ToLower(strings.TrimSpace(in.Email))

	// Check global email uniqueness using SystemDB
	var existingUser models.User
	err := database.SystemDB.Where("email = ?", email).First(&existingUser).Error
	if err == nil {
		return nil, ErrEmailExists
	}

	// Setup transaction
	tx := database.SystemDB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Defaults / fallbacks
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = "TRY"
	}
	locale := strings.ToLower(strings.TrimSpace(in.Locale))
	if locale == "" {
		locale = "tr"
	}
	timezone := strings.TrimSpace(in.Timezone)
	if timezone == "" {
		timezone = "Europe/Istanbul"
	}

	// Create Company
	companyID, _ := uuid.NewV7()
	company := models.Company{
		ID:                 companyID,
		Name:               in.CompanyName,
		ContactName:        in.Name,
		Slug:               uniqueSlug(tx, in.CompanyName),
		Email:              email,
		Phone:              strings.TrimSpace(in.Phone),
		Industry:           strings.TrimSpace(in.Industry),
		Country:            strings.TrimSpace(in.Country),
		City:               strings.TrimSpace(in.City),
		District:           strings.TrimSpace(in.District),
		Address:            strings.TrimSpace(in.Address),
		Currency:           currency,
		Locale:             locale,
		Timezone:           timezone,
		SubscriptionStatus: "trial",
		TrialEndsAt:        time.Now().AddDate(0, 0, 30),
	}

	if err := tx.Create(&company).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create User
	userID, _ := uuid.NewV7()
	passHash, err := utils.HashPassword(in.Password)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	user := models.User{
		ID:           userID,
		CompanyID:    &company.ID,
		Name:         in.Name,
		Email:        email,
		Phone:        strings.TrimSpace(in.Phone),
		PasswordHash: passHash,
		Role:         "admin",
		Locale:       locale,
		Timezone:     timezone,
		IsActive:     true,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		// Es zamanli kayitta DB unique (email) ihlali -> 409 mapping
		if isUniqueViolation(err) {
			return nil, ErrEmailExists
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

	// Initialize Default Product Categories
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

	// Initialize Default Expense Categories
	expCategories := []models.ExpenseCategory{
		{CompanyID: company.ID, Name: "Elektrik"},
		{CompanyID: company.ID, Name: "Su"},
		{CompanyID: company.ID, Name: "Doğalgaz"},
		{CompanyID: company.ID, Name: "İnternet"},
		{CompanyID: company.ID, Name: "Telefon"},
		{CompanyID: company.ID, Name: "Temizlik"},
		{CompanyID: company.ID, Name: "Kargo"},
		{CompanyID: company.ID, Name: "Yemek"},
	}
	for i := range expCategories {
		expCategories[i].ID, _ = uuid.NewV7()
	}
	if err := tx.Create(&expCategories).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Initialize Default Accounts
	// 1. Cash Accounts
	cashAccounts := []models.CashAccount{
		{CompanyID: company.ID, Code: "KAS-TL", Name: "Merkez Kasa (TRY)", Currency: "TRY", IsDefault: true},
		{CompanyID: company.ID, Code: "KAS-USD", Name: "Merkez Kasa (USD)", Currency: "USD"},
		{CompanyID: company.ID, Code: "KAS-EUR", Name: "Merkez Kasa (EUR)", Currency: "EUR"},
	}
	for i := range cashAccounts {
		cashAccounts[i].ID, _ = uuid.NewV7()
	}
	if err := tx.Create(&cashAccounts).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 2. Bank Accounts
	bankAccounts := []models.BankAccount{
		{CompanyID: company.ID, Code: "BNK-TL", Name: "Banka Hesabı (TRY)", Currency: "TRY"},
		{CompanyID: company.ID, Code: "BNK-USD", Name: "Banka Hesabı (USD)", Currency: "USD"},
		{CompanyID: company.ID, Code: "BNK-EUR", Name: "Banka Hesabı (EUR)", Currency: "EUR"},
	}
	for i := range bankAccounts {
		bankAccounts[i].ID, _ = uuid.NewV7()
	}
	if err := tx.Create(&bankAccounts).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 3. Default Currencies (selected currency marked default)
	currencies := []models.Currency{
		{CompanyID: company.ID, Name: "Türk Lirası", Symbol: "₺", Code: "TRY", ExchangeRate: decimal.NewFromInt(1), FormatPosition: "RightSpace", FormatThousandSep: ".", FormatDecimalSep: ",", FormatDecimals: 2, IsDefault: currency == "TRY"},
		{CompanyID: company.ID, Name: "Amerikan Doları", Symbol: "$", Code: "USD", ExchangeRate: decimal.NewFromInt(1), FormatPosition: "Left", FormatThousandSep: ",", FormatDecimalSep: ".", FormatDecimals: 2, IsDefault: currency == "USD"},
		{CompanyID: company.ID, Name: "Euro", Symbol: "€", Code: "EUR", ExchangeRate: decimal.NewFromInt(1), FormatPosition: "RightSpace", FormatThousandSep: ".", FormatDecimalSep: ",", FormatDecimals: 2, IsDefault: currency == "EUR"},
		{CompanyID: company.ID, Name: "İngiliz Sterlini", Symbol: "£", Code: "GBP", ExchangeRate: decimal.NewFromInt(1), FormatPosition: "Left", FormatThousandSep: ",", FormatDecimalSep: ".", FormatDecimals: 2, IsDefault: currency == "GBP"},
		{CompanyID: company.ID, Name: "Rus Rublesi", Symbol: "₽", Code: "RUB", ExchangeRate: decimal.NewFromInt(1), FormatPosition: "RightSpace", FormatThousandSep: ".", FormatDecimalSep: ",", FormatDecimals: 2, IsDefault: currency == "RUB"},
	}
	// Ensure at least one default if currency is none of the seeded ones
	hasDefault := false
	for i := range currencies {
		if currencies[i].IsDefault {
			hasDefault = true
		}
	}
	if !hasDefault {
		currencies[0].IsDefault = true
	}
	for i := range currencies {
		currencies[i].ID, _ = uuid.NewV7()
	}
	if err := tx.Create(&currencies).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Initialize Default Settings
	sId1, _ := uuid.NewV7()
	sId2, _ := uuid.NewV7()
	defaultCariGroups := `["Bireysel", "Kurumsal", "Kurum", "Fabrika", "Esnaf", "Şirket", "Diğer"]`

	settings := []models.Setting{
		{
			ID:        sId1,
			CompanyID: company.ID,
			Key:       "cari_groups",
			Value:     defaultCariGroups,
			Category:  "cari",
		},
		{
			ID:        sId2,
			CompanyID: company.ID,
			Key:       "quote_prefix",
			Value:     "PRO",
			Category:  "quote",
		},
	}
	if err := tx.Create(&settings).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &AuthResult{
		User:    user,
		Company: company,
	}, nil
}

func (s *AuthService) Login(identifier, password string) (*AuthResult, error) {
	identifier = strings.ToLower(strings.TrimSpace(identifier))

	var user models.User
	query := database.SystemDB.Where("email = ?", identifier)
	// If identifier looks like a phone (no @), match phone instead/also
	if !strings.Contains(identifier, "@") {
		digits := normalizePhone(identifier)
		query = database.SystemDB.Where("email = ? OR phone = ? OR phone = ?", identifier, identifier, digits)
	}
	if err := query.First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCreds
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return nil, ErrInvalidCreds
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	database.SystemDB.Model(&user).Update("last_login_at", now)

	// Fetch Company
	var company models.Company
	if user.CompanyID != nil {
		if err := database.SystemDB.First(&company, *user.CompanyID).Error; err != nil {
			return nil, err
		}
	}

	return s.generateSession(user, company)
}

func (s *AuthService) Refresh(rawRefreshToken string) (*AuthResult, error) {
	tokenHash := utils.HashToken(rawRefreshToken)

	var rt models.RefreshToken
	if err := database.SystemDB.Where("token_hash = ?", tokenHash).First(&rt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	if rt.RevokedAt != nil {
		return nil, ErrTokenRevoked
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	// Fetch User
	var user models.User
	if err := database.SystemDB.First(&user, rt.UserID).Error; err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Fetch Company
	var company models.Company
	if user.CompanyID != nil {
		if err := database.SystemDB.First(&company, *user.CompanyID).Error; err != nil {
			return nil, err
		}
	}

	// Rotate token: revoke old one
	now := time.Now()
	rt.RevokedAt = &now
	if err := database.SystemDB.Save(&rt).Error; err != nil {
		return nil, err
	}

	return s.generateSession(user, company)
}

func (s *AuthService) Logout(rawRefreshToken string) error {
	tokenHash := utils.HashToken(rawRefreshToken)

	var rt models.RefreshToken
	if err := database.SystemDB.Where("token_hash = ?", tokenHash).First(&rt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // Already logged out or invalid
		}
		return err
	}

	now := time.Now()
	rt.RevokedAt = &now
	return database.SystemDB.Save(&rt).Error
}

func (s *AuthService) Me(userID uuid.UUID) (*AuthResult, error) {
	var user models.User
	if err := database.SystemDB.First(&user, userID).Error; err != nil {
		return nil, err
	}

	var company models.Company
	if user.CompanyID != nil {
		if err := database.SystemDB.First(&company, *user.CompanyID).Error; err != nil {
			return nil, err
		}
	}

	return &AuthResult{
		User:        user,
		Company:     company,
		Permissions: loadPermissions(user),
	}, nil
}

func (s *AuthService) generateSession(user models.User, company models.Company) (*AuthResult, error) {
	// Sign Access Token
	// TTL config load or default 15m
	atTTL := 15 * time.Minute
	rtTTL := 7 * 24 * time.Hour

	// Get secret from environment
	secretStr := os.Getenv("JWT_SECRET")
	if secretStr == "" {
		return nil, errors.New("sistem anahtarı (JWT_SECRET) tanımlı değil")
	}
	secret := []byte(secretStr)

	at, err := utils.SignAccess(secret, atTTL, user.ID, company.ID, user.Role)
	if err != nil {
		return nil, err
	}

	// Generate and save Refresh Token
	rawRt, err := utils.GenerateRandomToken()
	if err != nil {
		return nil, err
	}

	rtID, _ := uuid.NewV7()
	rt := models.RefreshToken{
		ID:        rtID,
		UserID:    user.ID,
		TokenHash: utils.HashToken(rawRt),
		ExpiresAt: time.Now().Add(rtTTL),
	}

	if err := database.SystemDB.Create(&rt).Error; err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		Company:      company,
		Permissions:  loadPermissions(user),
		AccessToken:  at,
		RefreshToken: rawRt,
	}, nil
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	var sb strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// uniqueSlug, firma adindan benzersiz bir slug uretir.
// - Turkce/ozel karakterli adlarda slugify bos donerse "firma" tabanini kullanir.
// - Cakisma varsa -2, -3... ekler; cok nadir asiri durumda kisa uuid ekler.
func uniqueSlug(tx *gorm.DB, name string) string {
	base := slugify(name)
	if base == "" {
		base = "firma"
	}
	slug := base
	for i := 2; i <= 1000; i++ {
		var count int64
		tx.Model(&models.Company{}).Where("slug = ?", slug).Count(&count)
		if count == 0 {
			return slug
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
	return fmt.Sprintf("%s-%s", base, uuid.New().String()[:8])
}

// normalizePhone keeps only digits (drops +, spaces, dashes, parentheses).
func normalizePhone(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// isUniqueViolation, Postgres unique constraint (23505) ihlalini tespit eder.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "23505") ||
		strings.Contains(msg, "unique constraint")
}

func (s *AuthService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	var user models.User
	if err := database.SystemDB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if !utils.CheckPasswordHash(oldPassword, user.PasswordHash) {
		return ErrInvalidCreds
	}

	newHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return database.SystemDB.Model(&user).Update("password_hash", newHash).Error
}
