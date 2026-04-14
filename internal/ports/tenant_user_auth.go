package ports

import "github.com/google/uuid"

const (
	RoleUser        = "user"
	RoleTenantAdmin = "user_admin"
	RoleSystemAdmin = "system_admin"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type TenantUserAuthService interface {
	Login(tenantID uuid.UUID, email, password string) (*TokenPair, error)
	Refresh(refreshToken string) (*TokenPair, error)
}
