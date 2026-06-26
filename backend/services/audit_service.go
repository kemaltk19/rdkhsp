package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"radikal-hesap/models"
)

// WriteAuditLog inserts a snapshot record of an action to the audit logs table.
func WriteAuditLog(ctx context.Context, tx *gorm.DB, module string, recordID uuid.UUID, action string, userID uuid.UUID, summary string) error {
	var userName string
	var userRole string
	var pUserID *uuid.UUID

	if userID != uuid.Nil {
		var user models.User
		if err := tx.Select("name", "role").First(&user, "id = ?", userID).Error; err == nil {
			userName = user.Name
			userRole = user.Role
			pUserID = &userID
		} else {
			userName = "Unknown User"
		}
	} else {
		userName = "System"
	}

	var companyID uuid.UUID
	if ctx != nil {
		if compIDStr, ok := ctx.Value("company_id").(string); ok && compIDStr != "" {
			companyID, _ = uuid.Parse(compIDStr)
		}
	}

	return tx.Create(&models.AuditLog{
		ID:        uuid.New(),
		CompanyID: companyID,
		Module:    module,
		RecordID:  recordID,
		Action:    action,
		UserID:    pUserID,
		UserName:  userName,
		UserRole:  userRole,
		Summary:   summary,
		CreatedAt: time.Now(),
	}).Error
}
