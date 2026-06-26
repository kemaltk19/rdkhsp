package services

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"radikal-hesap/database"
	"radikal-hesap/models"
)

var (
	ErrCategoryExists     = errors.New("category_exists")
	ErrCategoryNotFound   = errors.New("category_not_found")
	ErrCategoryProtected  = errors.New("category_protected")
	ErrCategoryNameEmpty  = errors.New("category_name_empty")
	defaultCategorySlug   = "bilgi"
	slugSanitizeRe        = regexp.MustCompile(`[^a-z0-9]+`)
	turkishSlugReplacer   = strings.NewReplacer(
		"ç", "c", "ğ", "g", "ı", "i", "ö", "o", "ş", "s", "ü", "u",
		"Ç", "c", "Ğ", "g", "İ", "i", "Ö", "o", "Ş", "s", "Ü", "u",
	)
)

// categorySlugify, kategori adından slug üretir. auth_service'teki slugify'dan
// farkı: Türkçe karakterleri ASCII'ye çevirir (Eğitim -> egitim).
func categorySlugify(name string) string {
	s := turkishSlugReplacer.Replace(strings.TrimSpace(name))
	s = strings.ToLower(s)
	s = slugSanitizeRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

type AnnouncementService struct{}

func NewAnnouncementService() *AnnouncementService {
	return &AnnouncementService{}
}

func (s *AnnouncementService) Create(ctx context.Context, title, body, category string, targetPlanID *uuid.UUID, creatorID uuid.UUID) (*models.Announcement, error) {
	if category == "" {
		category = "bilgi"
	}

	ann := &models.Announcement{
		ID:           uuid.New(),
		Title:        title,
		Body:         body,
		Category:     category,
		TargetPlanID: targetPlanID,
		CreatedBy:    creatorID,
		CreatedAt:    time.Now(),
	}

	if err := database.SystemDB.Create(ann).Error; err != nil {
		return nil, err
	}
	return ann, nil
}

func (s *AnnouncementService) ListAll(ctx context.Context) ([]models.Announcement, error) {
	var list []models.Announcement
	if err := database.SystemDB.Order("created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *AnnouncementService) Delete(ctx context.Context, id uuid.UUID) error {
	return database.SystemDB.Delete(&models.Announcement{}, "id = ?", id).Error
}

func (s *AnnouncementService) ListForTenant(ctx context.Context, companyID uuid.UUID) ([]models.Announcement, error) {
	var comp models.Company
	if err := database.SystemDB.First(&comp, "id = ?", companyID).Error; err != nil {
		return nil, err
	}

	var list []models.Announcement
	query := database.SystemDB.Order("created_at desc")

	if comp.PlanID != nil {
		query = query.Where("target_plan_id IS NULL OR target_plan_id = ?", *comp.PlanID)
	} else {
		query = query.Where("target_plan_id IS NULL")
	}

	if err := query.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ── Kategori yönetimi ──

func (s *AnnouncementService) ListCategories(ctx context.Context) ([]models.AnnouncementCategory, error) {
	var list []models.AnnouncementCategory
	if err := database.SystemDB.Order("created_at asc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *AnnouncementService) CreateCategory(ctx context.Context, name string) (*models.AnnouncementCategory, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrCategoryNameEmpty
	}
	slug := categorySlugify(name)
	if slug == "" {
		return nil, ErrCategoryNameEmpty
	}

	var count int64
	database.SystemDB.Model(&models.AnnouncementCategory{}).Where("slug = ?", slug).Count(&count)
	if count > 0 {
		return nil, ErrCategoryExists
	}

	cat := &models.AnnouncementCategory{
		ID:        uuid.New(),
		Slug:      slug,
		Name:      name,
		CreatedAt: time.Now(),
	}
	if err := database.SystemDB.Create(cat).Error; err != nil {
		return nil, err
	}
	return cat, nil
}

// DeleteCategory bir kategoriyi siler. Varsayılan 'bilgi' silinemez; silinen
// kategoriyi kullanan mevcut duyurular 'bilgi'ye düşürülür (yetim kalmasın).
func (s *AnnouncementService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	var cat models.AnnouncementCategory
	if err := database.SystemDB.First(&cat, "id = ?", id).Error; err != nil {
		return ErrCategoryNotFound
	}
	if cat.Slug == defaultCategorySlug {
		return ErrCategoryProtected
	}

	return database.SystemDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Announcement{}).
			Where("category = ?", cat.Slug).
			Update("category", defaultCategorySlug).Error; err != nil {
			return err
		}
		return tx.Delete(&models.AnnouncementCategory{}, "id = ?", id).Error
	})
}
