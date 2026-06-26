package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)

var (
	ErrEmployeeEmailExists     = errors.New("employee_email_exists")
	ErrEmployeeNotFound        = errors.New("employee_not_found")
	ErrRoleNotFoundOrForbidden = errors.New("role_not_found_or_forbidden")
)

type EmployeeService struct{}

func NewEmployeeService() *EmployeeService {
	return &EmployeeService{}
}

type CreateEmployeeInput struct {
	Name                string     `json:"name" binding:"required,max=255"`
	Email               string     `json:"email" binding:"required,email,max=255"`
	Phone               string     `json:"phone" binding:"max=50"`
	Position            string     `json:"position" binding:"max=100"`
	Department          string     `json:"department" binding:"max=100"`
	HireDate            *time.Time `json:"hire_date"`
	GiveLoginPermission bool       `json:"give_login_permission"`
	Password            string     `json:"password"`
	RoleID              *uuid.UUID `json:"role_id"` // sadece GiveLoginPermission=true iken anlamlı
}

type UpdateEmployeeInput struct {
	Name                string     `json:"name" binding:"required,max=255"`
	Phone               string     `json:"phone" binding:"max=50"`
	Position            string     `json:"position" binding:"max=100"`
	Department          string     `json:"department" binding:"max=100"`
	HireDate            *time.Time `json:"hire_date"`
	IsActive            bool       `json:"is_active"`
	RoleID              *uuid.UUID `json:"role_id"`               // emp.UserID varsa rol değişikliği uygulanır
	GiveLoginPermission *bool      `json:"give_login_permission"` // nil = değiştirme; true + UserID yok ise yeni login oluşturulur
	Password            *string    `json:"password"`              // doluysa şifre güncellenir; UserID yoksa ve GiveLoginPermission=true ise ilk şifre olarak kullanılır
}

func (s *EmployeeService) List(ctx context.Context, page, limit int, search string, sort string, isActive *bool) ([]models.Employee, int64, error) {
	db := utils.GetDB(ctx, database.DB)
	var employees []models.Employee
	var total int64

	query := db.Model(&models.Employee{}).Preload("User.RoleRef")

	if search != "" {
		sClean := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", sClean, sClean)
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if sort != "" {
		query = query.Order(sort)
	} else {
		query = query.Order("created_at desc")
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&employees).Error; err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

func (s *EmployeeService) GetByID(ctx context.Context, id uuid.UUID) (*models.Employee, error) {
	db := utils.GetDB(ctx, database.DB)
	var emp models.Employee
	if err := db.Preload("User.RoleRef").First(&emp, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}
	return &emp, nil
}

func (s *EmployeeService) Create(ctx context.Context, in CreateEmployeeInput, userIDVal uuid.UUID) (*models.Employee, error) {
	db := utils.GetDB(ctx, database.DB)
	email := strings.ToLower(strings.TrimSpace(in.Email))

	// Check if email already used by a user globally
	var existingUser models.User
	if err := database.SystemDB.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, ErrEmployeeEmailExists
	}

	compIDStr := ctx.Value("company_id").(string)
	companyID, _ := uuid.Parse(compIDStr)

	var userID *uuid.UUID

	if in.GiveLoginPermission {
		if len(in.Password) < 8 {
			return nil, errors.New("password_too_short")
		}

		// RoleID verilmişse, bu rolün gerçekten bu şirkete ait olduğunu doğrula (cross-tenant koruması)
		if in.RoleID != nil {
			var count int64
			db.Model(&models.Role{}).Where("id = ? AND company_id = ?", *in.RoleID, companyID).Count(&count)
			if count == 0 {
				return nil, ErrRoleNotFoundOrForbidden
			}
		}

		uID, _ := uuid.NewV7()
		passHash, err := utils.HashPassword(in.Password)
		if err != nil {
			return nil, err
		}

		user := models.User{
			ID:           uID,
			CompanyID:    &companyID,
			Name:         in.Name,
			Email:        email,
			PasswordHash: passHash,
			Role:         "personel",
			RoleID:       in.RoleID,
			Locale:       "tr",
			IsActive:     true,
		}

		if err := db.Create(&user).Error; err != nil {
			return nil, err
		}
		userID = &uID
	}

	empID, _ := uuid.NewV7()
	employee := models.Employee{
		ID:         empID,
		CompanyID:  companyID,
		UserID:     userID,
		Name:       in.Name,
		Email:      email,
		Phone:      in.Phone,
		Position:   in.Position,
		Department: in.Department,
		HireDate:   in.HireDate,
		IsActive:   true,
	}

	if err := db.Create(&employee).Error; err != nil {
		return nil, err
	}

	if err := WriteAuditLog(ctx, db, "employee", employee.ID, "create", userIDVal, employee.Name); err != nil {
		return nil, err
	}

	return &employee, nil
}

func (s *EmployeeService) Update(ctx context.Context, id uuid.UUID, in UpdateEmployeeInput, userID uuid.UUID) (*models.Employee, error) {
	db := utils.GetDB(ctx, database.DB)

	// Fetch existing
	emp, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	emp.Name = in.Name
	emp.Phone = in.Phone
	emp.Position = in.Position
	emp.Department = in.Department
	emp.HireDate = in.HireDate
	emp.IsActive = in.IsActive

	if err := db.Save(emp).Error; err != nil {
		return nil, err
	}

	if emp.UserID != nil {
		var user models.User
		if err := db.First(&user, "id = ?", *emp.UserID).Error; err == nil {
			user.Name = in.Name
			user.IsActive = in.IsActive
			if in.RoleID != nil {
				var count int64
				db.Model(&models.Role{}).Where("id = ? AND company_id = ?", *in.RoleID, emp.CompanyID).Count(&count)
				if count > 0 {
					user.RoleID = in.RoleID
				}
			}
			if in.Password != nil && *in.Password != "" {
				if len(*in.Password) < 8 {
					return nil, errors.New("password_too_short")
				}
				passHash, err := utils.HashPassword(*in.Password)
				if err != nil {
					return nil, err
				}
				user.PasswordHash = passHash
			}
			if err := db.Save(&user).Error; err != nil {
				return nil, err
			}
		}
	} else if in.GiveLoginPermission != nil && *in.GiveLoginPermission {
		// Henüz login hesabı yok, sonradan giriş yetkisi veriliyor
		if in.Password == nil || len(*in.Password) < 8 {
			return nil, errors.New("password_too_short")
		}

		var existingUser models.User
		if err := database.SystemDB.Where("email = ?", emp.Email).First(&existingUser).Error; err == nil {
			return nil, ErrEmployeeEmailExists
		}

		if in.RoleID != nil {
			var count int64
			db.Model(&models.Role{}).Where("id = ? AND company_id = ?", *in.RoleID, emp.CompanyID).Count(&count)
			if count == 0 {
				return nil, ErrRoleNotFoundOrForbidden
			}
		}

		uID, _ := uuid.NewV7()
		passHash, err := utils.HashPassword(*in.Password)
		if err != nil {
			return nil, err
		}

		user := models.User{
			ID:           uID,
			CompanyID:    &emp.CompanyID,
			Name:         in.Name,
			Email:        emp.Email,
			PasswordHash: passHash,
			Role:         "personel",
			RoleID:       in.RoleID,
			Locale:       "tr",
			IsActive:     true,
		}
		if err := db.Create(&user).Error; err != nil {
			return nil, err
		}

		emp.UserID = &uID
		if err := db.Save(emp).Error; err != nil {
			return nil, err
		}
	}

	return emp, nil
}

func (s *EmployeeService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	db := utils.GetDB(ctx, database.DB)

	emp, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := WriteAuditLog(ctx, db, "employee", emp.ID, "delete", userID, emp.Name); err != nil {
		return err
	}

	// If linked user exists, delete refresh tokens & linked user
	if emp.UserID != nil {
		if err := db.Where("user_id = ?", *emp.UserID).Delete(&models.RefreshToken{}).Error; err != nil {
			return err
		}
		if err := db.Where("id = ?", *emp.UserID).Delete(&models.User{}).Error; err != nil {
			return err
		}
	}

	if err := db.Delete(emp).Error; err != nil {
		return err
	}

	return nil
}
