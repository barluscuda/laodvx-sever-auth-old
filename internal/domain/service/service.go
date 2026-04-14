package service

import (
	"github.com/barluscuda/laodvx-server-auth/config"
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

type Deps struct {
	Tenant              ports.TenantRepository
	TenantUser          ports.TenantUserRepository
	SystemUser          ports.SystemUserRepository
	SystemAdmin         ports.SystemAdminRepository
	RefreshToken        ports.RefreshTokenRepository
	PendingRegistration ports.PendingRegistrationStore
	LoginAttempt        ports.LoginAttemptRepository
	RegisterAttempt     ports.RegistrationAttemptRepository
	EmailSender         ports.EmailSender
}

func NewServices(d Deps, cfg config.Config) *Services {
	return &Services{
		TenantUser:     NewTenantUserService(d.TenantUser, d.PendingRegistration, d.RegisterAttempt, d.EmailSender, cfg.Email),
		Tenant:         NewTenantService(d.Tenant),
		TenantAdmin:    NewTenantAdminService(d.TenantUser),
		UserAuth:       NewAuthService(d.TenantUser, d.RefreshToken, d.LoginAttempt, cfg.JWT),
		SystemAdmin:    NewSystemAdminService(d.SystemUser, d.TenantUser, d.Tenant),
		SystemAuth:     NewSystemAuthService(d.SystemAdmin, d.RefreshToken, d.LoginAttempt, cfg.JWT),
		DevSystemAdmin: NewDevSystemAdminService(d.SystemAdmin),
	}
}
