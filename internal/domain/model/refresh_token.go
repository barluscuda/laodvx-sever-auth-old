package model

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	TID       uuid.UUID  `gorm:"type:uuid;primaryKey;column:tid" json:"-"`
	OwnerID   uuid.UUID  `gorm:"type:uuid;not null;index"        json:"-"`
	TenantID  *uuid.UUID `gorm:"type:uuid"                       json:"-"` // nil for system admin
	UsedAt    *time.Time `                                        json:"-"`
	ExpiresAt time.Time  `gorm:"not null"                        json:"-"`
	CreatedAt time.Time  `                                        json:"-"`
}
