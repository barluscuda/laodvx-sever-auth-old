package ports

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"

	"github.com/google/uuid"
)

type TenantAdminService interface {
	GetAll(tenantID uuid.UUID) ([]model.TenantUser, error)
	GetByEmail(tenantID uuid.UUID, email string) (*model.TenantUser, error)
	GetByID(tenantID uuid.UUID, id uuid.UUID) (*model.TenantUser, error)
}
