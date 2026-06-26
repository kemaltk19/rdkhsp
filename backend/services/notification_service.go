package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"radikal-hesap/database"
	"radikal-hesap/models"
	"radikal-hesap/utils"
)

var (
	ErrNotificationNotFound = errors.New("bildirim bulunamadı")
)

type NotificationService struct{}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

// CreateNotification creates a new database notification, supporting custom transaction block
func (s *NotificationService) CreateNotification(
	ctx context.Context,
	tx *gorm.DB,
	companyID uuid.UUID,
	notifType string,
	title string,
	message string,
	targetID *uuid.UUID,
	targetType string,
) error {
	db := utils.GetDB(ctx, database.DB)
	if tx != nil {
		db = tx
	}

	id, _ := uuid.NewV7()
	notif := models.Notification{
		ID:         id,
		CompanyID:  companyID,
		Type:       notifType,
		Title:      title,
		Message:    message,
		TargetID:   targetID,
		TargetType: targetType,
		IsRead:     false,
		CreatedAt:  time.Now(),
	}

	return db.Create(&notif).Error
}

// List returns active notifications for the current tenant
func (s *NotificationService) List(ctx context.Context, page, limit int) ([]models.Notification, int64, error) {
	tx := utils.GetDB(ctx, database.DB)

	var notifications []models.Notification
	var total int64

	dbQuery := tx.Model(&models.Notification{})

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := dbQuery.Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifications).Error
	if err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// MarkAsRead marks a single notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		var notif models.Notification
		if err := txTenant.First(&notif, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotificationNotFound
			}
			return err
		}

		now := time.Now()
		notif.IsRead = true
		notif.ReadAt = &now

		return txTenant.Save(&notif).Error
	})
}

// MarkAllAsRead marks all unread notifications of the tenant as read
func (s *NotificationService) MarkAllAsRead(ctx context.Context) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		txTenant := utils.GetDB(ctx, tx)

		now := time.Now()
		return txTenant.Model(&models.Notification{}).
			Where("is_read = false").
			Updates(map[string]interface{}{
				"is_read": true,
				"read_at": &now,
			}).Error
	})
}
