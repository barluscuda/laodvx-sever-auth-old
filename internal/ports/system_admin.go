package ports

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/google/uuid"
)

type SystemAdminRepository interface {
	Create(u *model.SystemAdmin) error
	GetAll() ([]model.SystemAdmin, error)
	GetByUsername(username string) (*model.SystemAdmin, error)
	GetByID(id uuid.UUID) (*model.SystemAdmin, error)
	Delete(id uuid.UUID) error
}

type DevSystemAdminService interface {
	Create(username, password string) (*model.SystemAdmin, error)
	GetAll() ([]model.SystemAdmin, error)
	GetByID(id uuid.UUID) (*model.SystemAdmin, error)
	Delete(id uuid.UUID) error
}
