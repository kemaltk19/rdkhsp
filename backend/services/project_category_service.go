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
)

var (
	ErrProjectCategoryNotFound = errors.New("project_category_not_found")
)

type ProjectCategoryService struct{}

func NewProjectCategoryService() *ProjectCategoryService {
	return &ProjectCategoryService{}
}

type ProjectCategoryInput struct {
	Name     string `json:"name" binding:"required,max=255"`
	Code     string `json:"code" binding:"max=50"`
	Color    string `json:"color" binding:"max=10"`
	IsActive bool   `json:"is_active"`
}

func (s *ProjectCategoryService) Create(ctx context.Context, in ProjectCategoryInput) (*models.ProjectCategory, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, err
	}

	category := models.ProjectCategory{
		ID:        uuid.New(),
		CompanyID: companyID,
		Name:      in.Name,
		Code:      in.Code,
		Color:     in.Color,
		IsActive:  in.IsActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	txTenant := utils.GetDB(ctx, database.DB)
	if err := txTenant.Create(&category).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return nil, fmt.Errorf("duplicate_project_category_code")
		}
		return nil, err
	}

	return &category, nil
}

func (s *ProjectCategoryService) Update(ctx context.Context, id uuid.UUID, in ProjectCategoryInput) (*models.ProjectCategory, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, err
	}

	var category models.ProjectCategory

	txTenant := utils.GetDB(ctx, database.DB)
	if err := txTenant.Where("id = ? AND company_id = ?", id, companyID).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectCategoryNotFound
		}
		return nil, err
	}

	category.Name = in.Name
	category.Code = in.Code
	category.Color = in.Color
	category.IsActive = in.IsActive
	category.UpdatedAt = time.Now()

	if err := txTenant.Save(&category).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return nil, fmt.Errorf("duplicate_project_category_code")
		}
		return nil, err
	}

	return &category, nil
}

func (s *ProjectCategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return err
	}

	txTenant := utils.GetDB(ctx, database.DB)
	var category models.ProjectCategory
	if err := txTenant.Where("id = ? AND company_id = ?", id, companyID).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectCategoryNotFound
		}
		return err
	}

	if err := txTenant.Delete(&category).Error; err != nil {
		return err
	}

	return nil
}

func (s *ProjectCategoryService) GetByID(ctx context.Context, id uuid.UUID) (*models.ProjectCategory, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, err
	}

	var category models.ProjectCategory
	txTenant := utils.GetDB(ctx, database.DB)

	if err := txTenant.Where("id = ? AND company_id = ?", id, companyID).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectCategoryNotFound
		}
		return nil, err
	}

	return &category, nil
}

func (s *ProjectCategoryService) List(ctx context.Context, page, limit int, query, sort string) ([]models.ProjectCategory, int64, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, 0, errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, 0, err
	}

	var categories []models.ProjectCategory
	var total int64
	txTenant := utils.GetDB(ctx, database.DB)

	q := txTenant.Where("company_id = ?", companyID)

	if query != "" {
		q = q.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ?", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%")
	}

	if err := q.Model(&models.ProjectCategory{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if sort == "" {
		sort = "-created_at"
	}
	orderBy := "created_at DESC"
	if strings.HasPrefix(sort, "-") {
		orderBy = strings.TrimPrefix(sort, "-") + " DESC"
	} else {
		orderBy = sort + " ASC"
	}

	offset := (page - 1) * limit
	if err := q.Order(orderBy).
		Offset(offset).
		Limit(limit).
		Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}
