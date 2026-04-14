package service

import (
	"github.com/barluscuda/laodvx-server-auth/config"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/repository"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"
)

type Services struct {
	TenantUser     ports.TenantUserService
	Tenant         ports.TenantService
	TenantAdmin    ports.TenantAdminService
	UserAuth       ports.TenantUserAuthService
	SystemAdmin    ports.SystemAdminService
	SystemAuth     ports.SystemAuthService
	DevSystemAdmin ports.DevSystemAdminService
}

func NewServices(repos *repository.Repository, cfg config.Config, emailSender ports.EmailSender) *Services {
	return &Services{
		TenantUser:     NewTenantUserService(repos.TenantUser, repos.PendingRegistration, emailSender, cfg.Email),
		Tenant:         NewTenantService(repos.Tenant),
		TenantAdmin:    NewTenantAdminService(repos.TenantUser),
		UserAuth:       NewAuthService(repos.TenantUser, repos.RefreshToken, repos.LoginAttempt, cfg.JWT),
		SystemAdmin:    NewSystemAdminService(repos.TenantUser),
		SystemAuth:     NewSystemAuthService(repos.SystemAdmin, repos.RefreshToken, repos.LoginAttempt, cfg.JWT),
		DevSystemAdmin: NewDevSystemAdminService(repos.SystemAdmin),
	}
}
