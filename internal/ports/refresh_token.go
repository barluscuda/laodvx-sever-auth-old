package ports

import (
	"time"

	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/google/uuid"
)

type RefreshTokenRepository interface {
	Create(rt *model.RefreshToken) error
	GetByTID(tid uuid.UUID) (*model.RefreshToken, error)
	MarkUsed(tid uuid.UUID, usedAt time.Time) error
	// ClaimToken atomically marks the token as used and returns it.
	// Returns ErrTokenAlreadyUsed if already claimed, ErrNotFound if missing or expired.
	ClaimToken(tid uuid.UUID) (*model.RefreshToken, error)
}
