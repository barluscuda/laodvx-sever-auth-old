package ports

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"

	"github.com/google/uuid"
)

type TenantUserRepository interface {
	Create(u *model.TenantUser) error
	GetByID(tenantID uuid.UUID, id uuid.UUID) (*model.TenantUser, error)
	GetByEmail(tenantID uuid.UUID, email string) (*model.TenantUser, error)
	GetAll(tenantID uuid.UUID) ([]model.TenantUser, error)
}

type TenantUserService interface {
	Create(tenantID uuid.UUID, email, password string) error
	GetByID(tenantID uuid.UUID, id uuid.UUID) (*model.TenantUser, error)
	GetByEmail(tenantID uuid.UUID, email string) (*model.TenantUser, error)
	VerifyEmail(tenantID uuid.UUID, email, otp string) error
	ResendVerification(tenantID uuid.UUID, email string) error
}
