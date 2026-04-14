package ports

import (
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"

	"github.com/google/uuid"
)

type TenantUserRepository interface {
	Create(u *model.TenantUser) error
	GetByID(tenantID, id uuid.UUID) (*model.TenantUser, error)
	GetByEmail(tenantID uuid.UUID, email string) (*model.TenantUser, error)
	GetAll(tenantID uuid.UUID) ([]model.TenantUser, error)
	SystemGetAll() ([]model.TenantUser, error)
	SystemGetAllByEmail(email string) ([]model.TenantUser, error)
	SystemGetByID(id uuid.UUID) (*model.TenantUser, error)
}

type TenantUserService interface {
	// Create stores a pending registration in cache and sends an OTP email.
	// The user is not written to the database until VerifyEmail succeeds.
	Create(tenantID uuid.UUID, email, password string) error
	GetByID(tenantID, id uuid.UUID) (*model.TenantUser, error)
	GetByEmail(tenantID uuid.UUID, email string) (*model.TenantUser, error)
	GetAllServer() ([]model.TenantUser, error)
	// VerifyEmail validates the OTP, creates the user in the database, and
	// removes the pending registration from cache.
	VerifyEmail(tenantID uuid.UUID, email, otp string) error
	// ResendVerification generates a new OTP and resends the verification email.
	// Always returns nil to avoid enumerating registered emails.
	ResendVerification(tenantID uuid.UUID, email string) error
}

type TenantAdminService interface {
	GetAll(tenantID uuid.UUID) ([]model.TenantUser, error)
	GetByEmail(tenantID uuid.UUID, email string) (*model.TenantUser, error)
	GetByID(tenantID, id uuid.UUID) (*model.TenantUser, error)
}

type SystemAdminService interface {
	GetAll() ([]model.TenantUser, error)
	GetAllByTenantID(tenantID uuid.UUID) ([]model.TenantUser, error)
	GetAllByEmail(email string) ([]model.TenantUser, error)
	GetByTenantAndEmail(tenantID uuid.UUID, email string) (*model.TenantUser, error)
	GetByID(id uuid.UUID) (*model.TenantUser, error)
}
