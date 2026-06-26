package utils

import (
	"context"

	"gorm.io/gorm"
)

type contextKey string

const TxKey contextKey = "db_tx"

// GetDB returns the database connection. If a transaction exists in the context,
// it returns that transaction. Otherwise, it returns the fallback connection.
func GetDB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(TxKey).(*gorm.DB); ok && tx != nil {
		return tx
	}
	return fallback
}
