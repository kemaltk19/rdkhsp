package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)

var (
	ErrRoleNotFound      = errors.New("role_not_found")
	ErrRoleInvalidModule = errors.New("invalid_module")
	ErrRoleInUse         = errors.New("role_in_use")
)

// allowedModules, rol sisteminin kapsadığı modüllerin whitelist'idir.
// employees ve settings kasıtlı olarak dışarıda — sadece admin/superadmin'e özel kalır.
var allowedModules = map[string]bool{
	"caris": true, "invoices": true, "payments": true,
	"expenses": true, "products": true, "reports": true,
}

type RoleService struct{}

func NewRoleService() *RoleService {
	return &RoleService{}
}

type RolePermissionInput struct {
	Module    string `json:"module" binding:"required"`
	CanCreate bool   `json:"can_create"`
	CanRead   bool   `json:"can_read"`
	CanUpdate bool   `json:"can_update"`
	CanDelete bool   `json:"can_delete"`
}

type RoleInput struct {
	Name        string                `json:"name" binding:"required"`
	Description string                `json:"description"`
	Permissions []RolePermissionInput `json:"permissions"`
}

func validateModules(perms []RolePermissionInput) error {
	for _, p := range perms {
		if !allowedModules[p.Module] {
			return ErrRoleInvalidModule
		}
	}
	return nil
}

func (s *RoleService) List(ctx context.Context) ([]models.Role, error) {
	tx := utils.GetDB(ctx, database.DB)
	var roles []models.Role
	if err := tx.Preload("Permissions").Order("created_at asc").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (s *RoleService) GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	tx := utils.GetDB(ctx, database.DB)
	var role models.Role
	if err := tx.Preload("Permissions").First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (s *RoleService) Create(ctx context.Context, in RoleInput, userID uuid.UUID) (*models.Role, error) {
	companyID, err := companyIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateModules(in.Permissions); err != nil {
		return nil, err
	}

	var role models.Role
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		role = models.Role{
			ID:          uuid.New(),
			CompanyID:   companyID,
			Name:        in.Name,
			Description: in.Description,
		}
		if err := txTenant.Create(&role).Error; err != nil {
			return err
		}

		for _, p := range in.Permissions {
			perm := models.RolePermission{
				ID:        uuid.New(),
				RoleID:    role.ID,
				Module:    p.Module,
				CanCreate: p.CanCreate,
				CanRead:   p.CanRead,
				CanUpdate: p.CanUpdate,
				CanDelete: p.CanDelete,
			}
			if err := txTenant.Create(&perm).Error; err != nil {
				return err
			}
		}
		if err := WriteAuditLog(ctx, txTenant, "role", role.ID, "create", userID, role.Name); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, role.ID)
}

func (s *RoleService) Update(ctx context.Context, id uuid.UUID, in RoleInput, userID uuid.UUID) (*models.Role, error) {
	if err := validateModules(in.Permissions); err != nil {
		return nil, err
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		var role models.Role
		if err := txTenant.First(&role, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRoleNotFound
			}
			return err
		}

		role.Name = in.Name
		role.Description = in.Description
		if err := txTenant.Save(&role).Error; err != nil {
			return err
		}

		// Mevcut izinleri sil, gönderileni yeniden oluştur — kısmi update karmaşıklığı yaratmaz.
		if err := txTenant.Where("role_id = ?", id).Delete(&models.RolePermission{}).Error; err != nil {
			return err
		}
		for _, p := range in.Permissions {
			perm := models.RolePermission{
				ID:        uuid.New(),
				RoleID:    id,
				Module:    p.Module,
				CanCreate: p.CanCreate,
				CanRead:   p.CanRead,
				CanUpdate: p.CanUpdate,
				CanDelete: p.CanDelete,
			}
			if err := txTenant.Create(&perm).Error; err != nil {
				return err
			}
		}
		if err := WriteAuditLog(ctx, txTenant, "role", id, "update", userID, role.Name); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, id)
}

func (s *RoleService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	tx := utils.GetDB(ctx, database.DB)

	var role models.Role
	if err := tx.First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}

	var userCount int64
	tx.Model(&models.User{}).Where("role_id = ?", id).Count(&userCount)
	if userCount > 0 {
		return ErrRoleInUse
	}

	if err := WriteAuditLog(ctx, tx, "role", id, "delete", userID, role.Name); err != nil {
		return err
	}

	if err := tx.Where("role_id = ?", id).Delete(&models.RolePermission{}).Error; err != nil {
		return err
	}
	return tx.Delete(&role).Error
}
