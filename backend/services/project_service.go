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
	ErrProjectNotFound = errors.New("project_not_found")
	ErrCariNotFoundForProject = errors.New("cari_not_found")
	ErrProjectNotEditable = errors.New("project_not_editable")
)

type ProjectService struct{}

func NewProjectService() *ProjectService {
	return &ProjectService{}
}

type ProjectInput struct {
	CariID      uuid.UUID   `json:"cari_id" binding:"required"`
	Name        string      `json:"name" binding:"required,max=255"`
	Description string      `json:"description" binding:"max=2000"`
	Code        string      `json:"code" binding:"max=50"`
	Status      string      `json:"status" binding:"required,oneof=planning in_progress on_hold completed cancelled"`
	CategoryID  *uuid.UUID  `json:"category_id"`
	EmployeeIDs []uuid.UUID `json:"employee_ids"`
	StartDate   time.Time   `json:"start_date" binding:"required"`
	EndDate     time.Time   `json:"end_date" binding:"required"`
	Budget      *float64    `json:"budget"`
	Note        string      `json:"note" binding:"max=2000"`
}

func (s *ProjectService) Create(ctx context.Context, in ProjectInput, createdBy uuid.UUID) (*models.Project, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, err
	}

	var project models.Project

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		// Verify Cari exists
		var cari models.Cari
		if err := txTenant.Where("id = ? AND company_id = ?", in.CariID, companyID).First(&cari).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCariNotFoundForProject
			}
			return err
		}

		// Generate project code if not provided
		code := in.Code
		if code == "" {
			var company models.Company
			if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", companyID).First(&company).Error; err != nil {
				return err
			}
			code = fmt.Sprintf("%s-%d", company.ProjectCodePrefix, company.ProjectCodeCounter)
			if err := tx.Model(&company).Update("project_code_counter", company.ProjectCodeCounter+1).Error; err != nil {
				return err
			}
		}

		// Create project
		project = models.Project{
			ID:          uuid.New(),
			CompanyID:   companyID,
			CariID:      in.CariID,
			Name:        in.Name,
			Description: in.Description,
			Code:        code,
			CategoryID:  in.CategoryID,
			Status:      in.Status,
			StartDate:   in.StartDate,
			EndDate:     in.EndDate,
			Budget:      in.Budget,
			Note:        in.Note,
			CreatedBy:   &createdBy,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := txTenant.Create(&project).Error; err != nil {
			if strings.Contains(err.Error(), "duplicate") {
				return fmt.Errorf("duplicate_project_code")
			}
			return err
		}

		// Associate employees if provided
		if len(in.EmployeeIDs) > 0 {
			var employees []models.Employee
			if err := txTenant.Where("id IN ? AND company_id = ?", in.EmployeeIDs, companyID).Find(&employees).Error; err != nil {
				return err
			}
			if err := txTenant.Model(&project).Association("Employees").Append(employees); err != nil {
				return err
			}
		}

		// Write audit log
		if err := WriteAuditLog(ctx, txTenant, "project", project.ID, "create", createdBy, project.Name); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (s *ProjectService) Update(ctx context.Context, id uuid.UUID, in ProjectInput, updatedBy uuid.UUID) (*models.Project, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, err
	}

	var project models.Project

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		if err := txTenant.Where("id = ? AND company_id = ?", id, companyID).First(&project).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProjectNotFound
			}
			return err
		}

		// Verify Cari exists
		var cari models.Cari
		if err := txTenant.Where("id = ? AND company_id = ?", in.CariID, companyID).First(&cari).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCariNotFoundForProject
			}
			return err
		}

		// Update project
		project.CariID = in.CariID
		project.Name = in.Name
		project.Description = in.Description
		project.Code = in.Code
		project.CategoryID = in.CategoryID
		project.Status = in.Status
		project.StartDate = in.StartDate
		project.EndDate = in.EndDate
		project.Budget = in.Budget
		project.Note = in.Note
		project.UpdatedBy = &updatedBy
		project.UpdatedAt = time.Now()

		if err := txTenant.Save(&project).Error; err != nil {
			if strings.Contains(err.Error(), "duplicate") {
				return fmt.Errorf("duplicate_project_code")
			}
			return err
		}

		// Update employees association
		if err := txTenant.Model(&project).Association("Employees").Clear(); err != nil {
			return err
		}
		if len(in.EmployeeIDs) > 0 {
			var employees []models.Employee
			if err := txTenant.Where("id IN ? AND company_id = ?", in.EmployeeIDs, companyID).Find(&employees).Error; err != nil {
				return err
			}
			if err := txTenant.Model(&project).Association("Employees").Append(employees); err != nil {
				return err
			}
		}

		// Write audit log
		if err := WriteAuditLog(ctx, txTenant, "project", project.ID, "update", updatedBy, project.Name); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (s *ProjectService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return err
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		var project models.Project
		if err := txTenant.Where("id = ? AND company_id = ?", id, companyID).First(&project).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProjectNotFound
			}
			return err
		}

		// Delete associations
		if err := txTenant.Model(&project).Association("Invoices").Clear(); err != nil {
			return err
		}
		if err := txTenant.Model(&project).Association("Quotes").Clear(); err != nil {
			return err
		}
		if err := txTenant.Model(&project).Association("Employees").Clear(); err != nil {
			return err
		}

		if err := txTenant.Delete(&project).Error; err != nil {
			return err
		}

		// Write audit log
		if err := WriteAuditLog(ctx, txTenant, "project", project.ID, "delete", userID, project.Name); err != nil {
			return err
		}

		return nil
	})
}

func (s *ProjectService) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, err
	}

	var project models.Project
	txTenant := utils.GetDB(ctx, database.DB)

	if err := txTenant.Where("id = ? AND company_id = ?", id, companyID).
		Preload("Invoices").
		Preload("Quotes").
		Preload("Employees").
		Preload("Category").
		Preload("CreatedByUser").
		Preload("UpdatedByUser").
		First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	return &project, nil
}

func (s *ProjectService) List(ctx context.Context, page, limit int, query, sort string, filters map[string]string) ([]models.Project, int64, error) {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return nil, 0, errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return nil, 0, err
	}

	var projects []models.Project
	var total int64
	txTenant := utils.GetDB(ctx, database.DB)

	q := txTenant.Where("company_id = ?", companyID)

	// Search by name or code
	if query != "" {
		q = q.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ?", "%"+strings.ToLower(query)+"%", "%"+strings.ToLower(query)+"%")
	}

	// Filters
	if status, ok := filters["status"]; ok && status != "" {
		q = q.Where("status = ?", status)
	}
	if cariID, ok := filters["cari_id"]; ok && cariID != "" {
		q = q.Where("cari_id = ?", cariID)
	}

	// Count total
	if err := q.Model(&models.Project{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sort
	if sort == "" {
		sort = "-created_at"
	}
	orderBy := "created_at DESC"
	if strings.HasPrefix(sort, "-") {
		orderBy = strings.TrimPrefix(sort, "-") + " DESC"
	} else {
		orderBy = sort + " ASC"
	}

	// Pagination
	offset := (page - 1) * limit
	if err := q.Order(orderBy).
		Offset(offset).
		Limit(limit).
		Preload("Employees").
		Preload("Category").
		Preload("CreatedByUser").
		Preload("UpdatedByUser").
		Find(&projects).Error; err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

func (s *ProjectService) AddInvoiceToProject(ctx context.Context, projectID, invoiceID uuid.UUID, userID uuid.UUID) error {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return err
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		// Verify project exists
		var project models.Project
		if err := txTenant.Where("id = ? AND company_id = ?", projectID, companyID).First(&project).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProjectNotFound
			}
			return err
		}

		// Verify invoice exists
		var invoice models.Invoice
		if err := txTenant.Where("id = ? AND company_id = ?", invoiceID, companyID).First(&invoice).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("invoice_not_found")
			}
			return err
		}

		// Add invoice to project
		if err := txTenant.Model(&project).Association("Invoices").Append(&invoice); err != nil {
			return err
		}

		// Update project's UpdatedAt and UpdatedBy
		if err := txTenant.Model(&project).Updates(map[string]interface{}{
			"updated_at": time.Now(),
			"updated_by": userID,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *ProjectService) AddQuoteToProject(ctx context.Context, projectID, quoteID uuid.UUID, userID uuid.UUID) error {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return err
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		// Verify project exists
		var project models.Project
		if err := txTenant.Where("id = ? AND company_id = ?", projectID, companyID).First(&project).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProjectNotFound
			}
			return err
		}

		// Verify quote exists
		var quote models.Quote
		if err := txTenant.Where("id = ? AND company_id = ?", quoteID, companyID).First(&quote).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("quote_not_found")
			}
			return err
		}

		// Add quote to project
		if err := txTenant.Model(&project).Association("Quotes").Append(&quote); err != nil {
			return err
		}

		// Update project's UpdatedAt and UpdatedBy
		if err := txTenant.Model(&project).Updates(map[string]interface{}{
			"updated_at": time.Now(),
			"updated_by": userID,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *ProjectService) RemoveInvoiceFromProject(ctx context.Context, projectID, invoiceID uuid.UUID, userID uuid.UUID) error {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return err
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		var project models.Project
		if err := txTenant.Where("id = ? AND company_id = ?", projectID, companyID).First(&project).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProjectNotFound
			}
			return err
		}

		if err := txTenant.Model(&project).Association("Invoices").Delete(&models.Invoice{}, "id = ?", invoiceID); err != nil {
			return err
		}

		if err := txTenant.Model(&project).Updates(map[string]interface{}{
			"updated_at": time.Now(),
			"updated_by": userID,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *ProjectService) RemoveQuoteFromProject(ctx context.Context, projectID, quoteID uuid.UUID, userID uuid.UUID) error {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return err
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		var project models.Project
		if err := txTenant.Where("id = ? AND company_id = ?", projectID, companyID).First(&project).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProjectNotFound
			}
			return err
		}

		if err := txTenant.Model(&project).Association("Quotes").Delete(&models.Quote{}, "id = ?", quoteID); err != nil {
			return err
		}

		if err := txTenant.Model(&project).Updates(map[string]interface{}{
			"updated_at": time.Now(),
			"updated_by": userID,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *ProjectService) AddEmployeeToProject(ctx context.Context, projectID, employeeID uuid.UUID, userID uuid.UUID) error {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return err
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		var project models.Project
		if err := txTenant.Where("id = ? AND company_id = ?", projectID, companyID).First(&project).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProjectNotFound
			}
			return err
		}

		var employee models.Employee
		if err := txTenant.Where("id = ? AND company_id = ?", employeeID, companyID).First(&employee).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("employee_not_found")
			}
			return err
		}

		if err := txTenant.Model(&project).Association("Employees").Append(&employee); err != nil {
			return err
		}

		if err := txTenant.Model(&project).Updates(map[string]interface{}{
			"updated_at": time.Now(),
			"updated_by": userID,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *ProjectService) RemoveEmployeeFromProject(ctx context.Context, projectID, employeeID uuid.UUID, userID uuid.UUID) error {
	companyIDStr := ctx.Value("company_id")
	if companyIDStr == nil {
		return errors.New("company_id not found in context")
	}
	companyID, err := uuid.Parse(companyIDStr.(string))
	if err != nil {
		return err
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		var project models.Project
		if err := txTenant.Where("id = ? AND company_id = ?", projectID, companyID).First(&project).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProjectNotFound
			}
			return err
		}

		if err := txTenant.Model(&project).Association("Employees").Delete(&models.Employee{}, "id = ?", employeeID); err != nil {
			return err
		}

		if err := txTenant.Model(&project).Updates(map[string]interface{}{
			"updated_at": time.Now(),
			"updated_by": userID,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}
