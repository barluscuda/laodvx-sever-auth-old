package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SystemAdmin struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement"      json:"-"`
	UUID      uuid.UUID      `gorm:"uniqueIndex;type:uuid;not null" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:100;not null"  json:"username"`
	Password  string         `gorm:"size:255;not null"              json:"-"`
	CreatedAt time.Time      `                                      json:"created_at"`
	UpdatedAt time.Time      `                                      json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                          json:"-"`
}

func (s *SystemAdmin) BeforeCreate(tx *gorm.DB) error {
	s.UUID = uuid.New()
	return nil
}
