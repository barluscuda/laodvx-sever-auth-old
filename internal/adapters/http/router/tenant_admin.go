package router

import (
	"github.com/barluscuda/laodvx-server-auth/config"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/handler"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/middleware"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/gin-gonic/gin"
)

// SetupTenantAdminRouter mounts tenant-scoped user-admin routes under /api/admin
func SetupTenantAdminRouter(r *gin.RouterGroup, cfg config.Config, userHandler *handler.TenantAdminHandler) {
	r.Use(middleware.RequireAuth(cfg.JWT.AccessKeys, ports.RoleTenantAdmin))

	u := r.Group("/user")
	{
		u.GET("", userHandler.GetAll)
		u.GET("/:id", userHandler.GetByID)
	}
}

// SetupSystemAdminRouter mounts system-admin routes under /system/api (main port)
func SetupSystemAdminRouter(r *gin.RouterGroup, cfg config.Config, h *handler.SystemAdminHandler, systemAuthHandler *handler.SystemAuthHandler) {
	// Public: system auth
	auth := r.Group("/auth")
	{
		auth.POST("/login", systemAuthHandler.Login)
		auth.POST("/refresh", systemAuthHandler.Refresh)
	}

	// Protected: system admin operations
	protected := r.Group("")
	protected.Use(middleware.RequireAuth(cfg.JWT.AccessKeys, ports.RoleSystemAdmin))
	{
		s := protected.Group("/tenant")
		{
			s.POST("", h.Create)
			s.GET("", h.GetAll)
			s.GET("/:id", h.GetByID)
			s.PUT("/:id", h.Update)
			s.DELETE("/:id", h.Delete)
		}

		u := protected.Group("/user")
		{
			u.GET("", h.GetAllUsers)
			u.GET("/:id", h.GetUserByID)
		}
	}
}
