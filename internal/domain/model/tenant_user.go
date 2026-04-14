package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TenantUser struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement"                         json:"-"`
	UUID      uuid.UUID      `gorm:"uniqueIndex;type:uuid;not null"                   json:"id"`
	TenantID  uuid.UUID      `gorm:"uniqueIndex:idx_tenant_email;type:uuid;not null"  json:"-"`
	Email     string         `gorm:"uniqueIndex:idx_tenant_email;size:255;not null"   json:"email"`
	Password  string         `gorm:"size:255;not null"                               json:"-"`
	Role      string         `gorm:"size:50;not null;default:'user'"                 json:"role"`
	CreatedAt time.Time      `                                                       json:"created_at"`
	UpdatedAt time.Time      `                                                       json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                           json:"-"`
}

func (u *TenantUser) BeforeCreate(tx *gorm.DB) error {
	u.UUID = uuid.New()
	return nil
}
