package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
	"strings"
)

var (
	ErrCompanyNotFound = errors.New("company_not_found")
)

type SettingsService struct{}

func NewSettingsService() *SettingsService {
	return &SettingsService{}
}

type UpdateCompanyInput struct {
	Name        string `json:"name" binding:"required"`
	ContactName string `json:"contact_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Landline    string `json:"landline"`
	Fax         string `json:"fax"`
	TaxOffice   string `json:"tax_office"`
	TaxNumber   string `json:"tax_number"`
	Industry    string `json:"industry"`
	Country     string `json:"country"`
	City        string `json:"city"`
	District    string `json:"district"`
	Address     string `json:"address"`
	LogoURL     string `json:"logo_url"`
	Currency    string `json:"currency" binding:"required"`
	Locale      string `json:"locale" binding:"required"`
	Timezone    string `json:"timezone"`
}

type SaveSettingInput struct {
	Key      string `json:"key" binding:"required"`
	Value    string `json:"value"`
	Category string `json:"category"`
}

func (s *SettingsService) GetCompanyProfile(ctx context.Context) (*models.Company, error) {
	compIDStr := ctx.Value("company_id").(string)
	companyID, err := uuid.Parse(compIDStr)
	if err != nil {
		return nil, ErrCompanyNotFound
	}

	var comp models.Company
	db := utils.GetDB(ctx, database.DB)
	if err := db.First(&comp, "id = ?", companyID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCompanyNotFound
		}
		return nil, err
	}

	// Eğer ContactName boş ise, bu eski kayıttır, admin kullanıcısının ismini çekip gösterelim
	if comp.ContactName == "" {
		var admin models.User
		if err := db.First(&admin, "company_id = ? AND role = ?", companyID, "admin").Error; err == nil {
			comp.ContactName = admin.Name
		}
	}

	return &comp, nil
}

func (s *SettingsService) UpdateCompanyProfile(ctx context.Context, in UpdateCompanyInput) (*models.Company, error) {
	compIDStr := ctx.Value("company_id").(string)
	companyID, _ := uuid.Parse(compIDStr)

	db := utils.GetDB(ctx, database.DB)
	var comp models.Company
	if err := db.First(&comp, "id = ?", companyID).Error; err != nil {
		return nil, ErrCompanyNotFound
	}

	comp.Name = in.Name
	comp.ContactName = in.ContactName
	comp.Email = in.Email
	comp.Phone = in.Phone
	comp.Landline = in.Landline
	comp.Fax = in.Fax
	comp.TaxOffice = in.TaxOffice
	comp.TaxNumber = in.TaxNumber
	comp.Industry = in.Industry
	comp.Country = in.Country
	comp.City = in.City
	comp.District = in.District
	comp.Address = in.Address
	comp.LogoURL = in.LogoURL
	comp.Currency = in.Currency
	comp.Locale = in.Locale
	if in.Timezone != "" {
		comp.Timezone = in.Timezone
	}
	comp.UpdatedAt = time.Now()

	if err := db.Save(&comp).Error; err != nil {
		return nil, err
	}

	// Base para birimi tek kaynaktır (Company.Currency). Currency tablosundaki
	// IsDefault bayrağını ve base kurunu (1.0) buna göre senkronla.
	if err := SyncDefaultCurrency(db, comp.Currency); err != nil {
		return nil, err
	}

	return &comp, nil
}

func (s *SettingsService) GetSetting(ctx context.Context, key string) (*models.Setting, error) {
	db := utils.GetDB(ctx, database.DB)
	var setting models.Setting
	if err := db.First(&setting, "key = ?", key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil when not found so frontend can use default value
		}
		return nil, err
	}
	return &setting, nil
}

func (s *SettingsService) SaveSetting(ctx context.Context, in SaveSettingInput) (*models.Setting, error) {
	db := utils.GetDB(ctx, database.DB)
	compIDStr := ctx.Value("company_id").(string)
	companyID, _ := uuid.Parse(compIDStr)

	var setting models.Setting
	err := db.First(&setting, "key = ?", in.Key).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new
			settID, _ := uuid.NewV7()
			category := in.Category
			if category == "" {
				category = "general"
			}
			setting = models.Setting{
				ID:        settID,
				CompanyID: companyID,
				Key:       in.Key,
				Value:     in.Value,
				Category:  category,
			}
			if err := db.Create(&setting).Error; err != nil {
				return nil, err
			}

			// If it's a prefix setting, reset its sequence counter
			if strings.HasSuffix(in.Key, "_prefix") {
				seqKey := strings.TrimSuffix(in.Key, "_prefix")
				db.Exec("UPDATE number_sequences SET last_no = 0 WHERE company_id = ? AND key = ?", companyID, seqKey)
			}

			return &setting, nil
		}
		return nil, err
	}

	// Update existing
	setting.Value = in.Value
	if in.Category != "" {
		setting.Category = in.Category
	}
	setting.UpdatedAt = time.Now()

	if err := db.Save(&setting).Error; err != nil {
		return nil, err
	}

	// If it's a prefix setting, reset its sequence counter
	if strings.HasSuffix(in.Key, "_prefix") {
		seqKey := strings.TrimSuffix(in.Key, "_prefix")
		db.Exec("UPDATE number_sequences SET last_no = 0 WHERE company_id = ? AND key = ?", companyID, seqKey)
	}

	return &setting, nil
}

func (s *SettingsService) ListSettings(ctx context.Context, category string) ([]models.Setting, error) {
	db := utils.GetDB(ctx, database.DB)
	var settings []models.Setting

	query := db.Model(&models.Setting{})
	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *SettingsService) UpdateEnabledModules(ctx context.Context, enabledModules []string) (*models.Company, error) {
	compIDStr := ctx.Value("company_id").(string)
	companyID, _ := uuid.Parse(compIDStr)

	db := utils.GetDB(ctx, database.DB)
	var comp models.Company
	if err := db.First(&comp, "id = ?", companyID).Error; err != nil {
		return nil, ErrCompanyNotFound
	}

	// Double check: modules must be within their Plan features
	if comp.PlanID != nil {
		var plan models.Plan
		if err := database.SystemDB.First(&plan, "id = ?", *comp.PlanID).Error; err == nil {
			if plan.Features != "" {
				var planFeatures []string
				if err := json.Unmarshal([]byte(plan.Features), &planFeatures); err == nil {
					featureMap := make(map[string]bool)
					for _, f := range planFeatures {
						featureMap[f] = true
					}
					var filtered []string
					for _, m := range enabledModules {
						if featureMap[m] {
							filtered = append(filtered, m)
						}
					}
					enabledModules = filtered
				}
			}
		}
	}

	// Let's do it cleaner by using package-level imports or doing it simple. Go json.Marshal.
	js, err := json.Marshal(enabledModules)
	if err != nil {
		return nil, err
	}

	jsStr := string(js)
	comp.EnabledModules = &jsStr
	comp.UpdatedAt = time.Now()

	if err := db.Save(&comp).Error; err != nil {
		return nil, err
	}
	return &comp, nil
}
