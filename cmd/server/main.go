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
)

func main() {
	cfg := config.Get()

	db := database.Connect(cfg.Database)
	db.AutoMigrate(
		&model.Tenant{},
		&model.TenantUser{},
		&model.SystemAdmin{},
		&model.RefreshToken{},
	)

	rdb := redis.Connect(cfg.Redis)

	var emailSender ports.EmailSender
	if cfg.Email.SMTPHost != "" {
		emailSender = emailadapter.NewSMTPSender(
			cfg.Email.SMTPHost,
			cfg.Email.SMTPPort,
			cfg.Email.SMTPUsername,
			cfg.Email.SMTPPassword,
			cfg.Email.FromAddress,
		)
	} else {
		emailSender = emailadapter.NoopSender{}
	}

	repos := repository.New(db, rdb, cfg.Redis.CacheTTL, cfg.RateLimit)
	svcs := svc.NewServices(repos, cfg, emailSender)

	server.New(cfg, repos.Tenant, server.Handlers{
		TenantUser:     httphandler.NewTenantUserHandler(svcs.TenantUser),
		Auth:           httphandler.NewAuthHandler(svcs.UserAuth),
		SystemAdmin:    httphandler.NewSystemAdminHandler(svcs.Tenant, svcs.SystemAdmin),
		SystemAuth:     httphandler.NewSystemAuthHandler(svcs.SystemAuth),
		TenantAdmin:    httphandler.NewTenantAdminHandler(svcs.TenantAdmin),
		DevSystemAdmin: httphandler.NewDevSystemAdminHandler(svcs.DevSystemAdmin),
	}).Run()
}
