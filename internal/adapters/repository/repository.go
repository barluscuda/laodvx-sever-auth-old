package repository

import "github.com/barluscuda/laodvx-server-auth/internal/ports"

// Repository bundles the DB-backed repositories. Redis-native stores and
// cache wrappers are composed by the caller (see cmd/server).
type Repository struct {
	Tenant       ports.TenantRepository
	TenantUser   ports.TenantUserRepository
	SystemUser   ports.SystemUserRepository
	SystemAdmin  ports.SystemAdminRepository
	RefreshToken ports.RefreshTokenRepository
}
