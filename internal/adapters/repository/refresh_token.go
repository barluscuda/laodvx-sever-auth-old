package repository

import (
	"errors"
	"time"

	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) ports.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(rt *model.RefreshToken) error {
	return r.db.Create(rt).Error
}

func (r *refreshTokenRepository) GetByTID(tid uuid.UUID) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	if err := r.db.First(&rt, "tid = ?", tid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ports.ErrNotFound
		}
		return nil, err
	}
	return &rt, nil
}

func (r *refreshTokenRepository) MarkUsed(tid uuid.UUID, usedAt time.Time) error {
	return r.db.Model(&model.RefreshToken{}).Where("tid = ?", tid).Update("used_at", usedAt).Error
}

func (r *refreshTokenRepository) ClaimToken(tid uuid.UUID) (*model.RefreshToken, error) {
	now := time.Now()
	rt := model.RefreshToken{}
	result := r.db.Model(&rt).
		Clauses(clause.Returning{}).
		Where("tid = ? AND used_at IS NULL AND expires_at > ?", tid, now).
		Update("used_at", now)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		// Could be already used, expired, or not found — all treated as invalid
		if err := r.db.First(&rt, "tid = ?", tid).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ports.ErrNotFound
			}
			return nil, err
		}
		if rt.UsedAt != nil {
			return nil, ports.ErrTokenAlreadyUsed
		}
		return nil, ports.ErrTokenExpired
	}
	return &rt, nil
}
