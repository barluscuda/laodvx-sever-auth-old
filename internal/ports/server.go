package ports

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/google/uuid"
)

type TenantRepository interface {
	Create(t *model.Tenant) error
	GetAll() ([]model.Tenant, error)
	GetByID(id uuid.UUID) (*model.Tenant, error)
	GetByName(name string) (*model.Tenant, error)
	Update(t *model.Tenant) error
	Delete(id uuid.UUID) error
	ExistsByID(id uuid.UUID) (bool, error)
}

type TenantService interface {
	Create(name string) (*model.Tenant, error)
	GetAll() ([]model.Tenant, error)
	GetByID(id uuid.UUID) (*model.Tenant, error)
	Update(id uuid.UUID, name string) (*model.Tenant, error)
	Delete(id uuid.UUID) error
}
