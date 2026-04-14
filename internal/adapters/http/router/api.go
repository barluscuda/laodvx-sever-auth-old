package router

import (
	"github.com/barluscuda/laodvx-server-auth/config"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/handler"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/middleware"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/gin-gonic/gin"
)

func SetupAPIRouter(r *gin.RouterGroup, cfg config.Config, userHandler *handler.TenantUserHandler, authHandler *handler.AuthHandler) {
	r.GET("/ping", handler.Ping)

	// All /api routes require an X-Tenant-Id header (already validated by
	// EnforceRouteScope + ExtractTenantIDFromHeader upstream). RequireTenant
	// lets handlers call MustTenantUUID without per-endpoint null checks.
	r.Use(middleware.RequireTenant())

	// Public: registration
	r.POST("/user", userHandler.Create)

	// Public: auth
	auth := r.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/verify-email", userHandler.VerifyEmail)
	}

	// Public: resend verification
	r.POST("/user/resend-verification", userHandler.ResendVerification)

	// Protected: self-service
	me := r.Group("/user")
	me.Use(middleware.RequireAuth(cfg.JWT.AccessKeys, ports.RoleUser, ports.RoleTenantAdmin))
	{
		me.GET("/me", userHandler.GetMe)
	}
}
