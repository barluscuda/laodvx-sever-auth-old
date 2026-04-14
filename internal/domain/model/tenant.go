package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tenant struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement"              json:"-"`
	UUID       uuid.UUID      `gorm:"uniqueIndex;type:uuid;not null"         json:"id"`
	TenantName string         `gorm:"uniqueIndex;size:100;not null"          json:"tenant_name"`
	Label      string         `gorm:"size:255;not null"                      json:"label"`
	Plan       string         `gorm:"size:20;not null;default:'neo'"         json:"plan"`
	Banned     bool           `gorm:"not null;default:false"                 json:"banned"`
	CreatedAt  time.Time      `                                              json:"created_at"`
	UpdatedAt  time.Time      `                                              json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index"                                  json:"-"`
}

func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	t.UUID = uuid.New()
	return nil
}
