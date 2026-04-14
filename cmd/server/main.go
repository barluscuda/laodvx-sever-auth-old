package main

import (
	"github.com/barluscuda/laodvx-server-auth/config"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/database"
	emailadapter "github.com/barluscuda/laodvx-server-auth/internal/adapters/email"
	httphandler "github.com/barluscuda/laodvx-server-auth/internal/adapters/http/handler"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/redis"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/repository"
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	svc "github.com/barluscuda/laodvx-server-auth/internal/domain/service"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"
	"github.com/barluscuda/laodvx-server-auth/internal/server"

	"gorm.io/gorm"
)

func main() {
	cfg := config.Get()

	db := database.Connect(cfg.Database)
	autoMigrate(db)

	repos := repository.New(db, redis.Connect(cfg.Redis), cfg.Redis.CacheTTL, cfg.RateLimit)
	services := svc.NewServices(repos, cfg, newEmailSender(cfg))

	server.New(cfg, repos.Tenant, newHandlers(services)).Run()
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
