package repository

import (
	"errors"
	"strings"

	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"gorm.io/gorm"
)

func firstOrNotFound(db *gorm.DB, dest any, query string, args ...any) error {
	if err := db.Where(query, args...).First(dest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.ErrNotFound
		}
		return err
	}
	return nil
}

func existsBy(db *gorm.DB, dest any, query string, args ...any) (bool, error) {
	result := db.Where(query, args...).Limit(1).Find(dest)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func requireAffected(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ports.ErrNotFound
	}
	return nil
}

// isDuplicateKeyErr centralises the fragile driver-specific string match so we
// only have to update it in one place if the driver or wrapping changes.
func isDuplicateKeyErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "UNIQUE constraint")
}
