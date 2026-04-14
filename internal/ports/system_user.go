package ports

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"

	"github.com/google/uuid"
)

// SystemUserRepository extends TenantUserRepository with system-wide (cross-tenant) operations.
type SystemUserRepository interface {
	TenantUserRepository
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
