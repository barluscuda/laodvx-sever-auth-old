package ports

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/google/uuid"
)

type TenantRepository interface {
	Create(t *model.Tenant) error
	GetAll() ([]model.Tenant, error)
	GetByTenantName(tenantName string) (*model.Tenant, error)
	Update(t *model.Tenant) error
	Delete(tenantName string) error
	ExistsByName(tenantName string) (bool, error)
	ExistsByUUID(id uuid.UUID) (bool, error)
	Ban(tenantName string) error
}

type TenantService interface {
	Create(tenantName, label, plan string) (*model.Tenant, error)
	GetAll() ([]model.Tenant, error)
	GetByTenantName(tenantName string) (*model.Tenant, error)
	Update(tenantName, label string) (*model.Tenant, error)
	Delete(tenantName string) error
	Ban(tenantName string) error
}
