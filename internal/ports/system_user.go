package ports

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"

	"github.com/google/uuid"
)

// SystemUserRepository exposes cross-tenant operations over users.
// Tenant-scoped CRUD lives in TenantUserRepository.
type SystemUserRepository interface {
	SystemGetAll() ([]model.TenantUser, error)
	SystemGetAllByEmail(email string) ([]model.TenantUser, error)
	SystemGetByID(id uuid.UUID) (*model.TenantUser, error)
	SystemSetRole(id uuid.UUID, role string) error
}

type SystemAdminService interface {
	GetAll() ([]model.TenantUser, error)
	GetAllByTenantName(tenantName string) ([]model.TenantUser, error)
	GetAllByEmail(email string) ([]model.TenantUser, error)
	GetByTenantAndEmail(tenantName string, email string) (*model.TenantUser, error)
	GetByID(id uuid.UUID) (*model.TenantUser, error)
	SetTenantAdmin(tenantName string, userID uuid.UUID) error
}
