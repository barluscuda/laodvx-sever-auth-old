package repository

import (
	"time"

	"github.com/barluscuda/laodvx-server-auth/config"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repository struct {
	Tenant              ports.TenantRepository
	TenantUser          ports.TenantUserRepository
	PendingRegistration ports.PendingRegistrationStore
	SystemAdmin         ports.SystemAdminRepository
	RefreshToken        ports.RefreshTokenRepository
	LoginAttempt        ports.LoginAttemptRepository
}

func New(db *gorm.DB, rdb *redis.Client, cacheTTL time.Duration, rateLimitCfg config.RateLimitConfig) *Repository {
	return &Repository{
		Tenant:              NewCachedTenantRepository(NewTenantRepository(db), rdb, cacheTTL),
		TenantUser:          NewCachedTenantUserRepository(NewTenantUserRepository(db), rdb, cacheTTL),
		PendingRegistration: NewPendingRegistrationStore(rdb),
		SystemAdmin:         NewSystemAdminRepository(db),
		RefreshToken:        NewRefreshTokenRepository(db),
		LoginAttempt:        NewLoginAttemptRepository(rdb, rateLimitCfg),
	}
}
