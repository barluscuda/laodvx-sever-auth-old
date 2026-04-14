package repository

import (
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
	if err := firstOrNotFound(r.db, &rt, "tid = ?", tid); err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *refreshTokenRepository) MarkUsed(tid uuid.UUID, usedAt time.Time) error {
	return r.db.Model(&model.RefreshToken{}).Where("tid = ?", tid).Update("used_at", usedAt).Error
}

func (r *refreshTokenRepository) ClaimToken(tid uuid.UUID) (*model.RefreshToken, error) {
	now := time.Now()
	var rt model.RefreshToken
	result := r.db.Model(&rt).
		Clauses(clause.Returning{}).
		Where("tid = ? AND used_at IS NULL AND expires_at > ?", tid, now).
		Update("used_at", now)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected > 0 {
		return &rt, nil
	}
	return nil, r.claimFailureReason(tid)
}

func (r *refreshTokenRepository) claimFailureReason(tid uuid.UUID) error {
	var rt model.RefreshToken
	if err := firstOrNotFound(r.db, &rt, "tid = ?", tid); err != nil {
		return err
	}
	if rt.UsedAt != nil {
		return ports.ErrTokenAlreadyUsed
	}
	return ports.ErrTokenExpired
}
