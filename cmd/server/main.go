package main

import (
	"github.com/barluscuda/laodvx-server-auth/config"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/cache"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/cachedrepo"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/database"
	emailadapter "github.com/barluscuda/laodvx-server-auth/internal/adapters/email"
	httphandler "github.com/barluscuda/laodvx-server-auth/internal/adapters/http/handler"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/redis"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/redisstore"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/repository"
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	svc "github.com/barluscuda/laodvx-server-auth/internal/domain/service"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"
	"github.com/barluscuda/laodvx-server-auth/internal/server"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Get()

	db := database.Connect(cfg.Database)
	autoMigrate(db)

	rdb := redis.Connect(cfg.Redis)

	deps := buildDeps(db, rdb, cfg, newEmailSender(cfg))
	services := svc.NewServices(deps, cfg)

	server.New(cfg, deps.Tenant, newHandlers(services)).Run()
}

func buildDeps(db *gorm.DB, rdb *goredis.Client, cfg config.Config, emailSender ports.EmailSender) svc.Deps {
	ttl := cfg.Redis.CacheTTL

	tenantRepo := repository.NewTenantRepository(db)
	tenantUserRepo := repository.NewTenantUserRepository(db)

	tenantCache := cache.NewTenantCache(rdb, ttl)
	tenantUserCache := cache.NewTenantUserCache(rdb, ttl)

	return svc.Deps{
		Tenant:              cachedrepo.NewTenant(tenantRepo, tenantCache),
		TenantUser:          cachedrepo.NewTenantUser(tenantUserRepo, tenantUserCache),
		SystemUser:          repository.NewSystemGlobalUserRepository(db),
		SystemAdmin:         repository.NewSystemAdminRepository(db),
		RefreshToken:        repository.NewRefreshTokenRepository(db),
		PendingRegistration: redisstore.NewPendingRegistrationStore(rdb),
		LoginAttempt:        redisstore.NewLoginAttemptRepository(rdb, cfg.RateLimit),
		EmailSender:         emailSender,
	}
}

func autoMigrate(db *gorm.DB) {
	db.AutoMigrate(
		&model.Tenant{},
		&model.TenantUser{},
		&model.SystemAdmin{},
		&model.RefreshToken{},
	)
}

func newEmailSender(cfg config.Config) ports.EmailSender {
	if cfg.Email.SMTPHost == "" {
		return emailadapter.NoopSender{}
	}
	return emailadapter.NewSMTPSender(
		cfg.Email.SMTPHost,
		cfg.Email.SMTPPort,
		cfg.Email.SMTPUsername,
		cfg.Email.SMTPPassword,
		cfg.Email.FromAddress,
	)
}

func newHandlers(s *svc.Services) server.Handlers {
	return server.Handlers{
		TenantUser:     httphandler.NewTenantUserHandler(s.TenantUser),
		Auth:           httphandler.NewAuthHandler(s.UserAuth),
		SystemAdmin:    httphandler.NewSystemAdminHandler(s.Tenant, s.SystemAdmin),
		SystemAuth:     httphandler.NewSystemAuthHandler(s.SystemAuth),
		TenantAdmin:    httphandler.NewTenantAdminHandler(s.TenantAdmin),
		DevSystemAdmin: httphandler.NewDevSystemAdminHandler(s.DevSystemAdmin),
	}
}
