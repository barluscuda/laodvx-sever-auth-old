package router

import (
	"github.com/barluscuda/laodvx-server-auth/config"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/handler"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/middleware"
	"github.com/gin-contrib/pprof"

	"github.com/gin-gonic/gin"
)

func SetupDevRouter(r *gin.Engine, cfg config.Config, devSystemAdmin *handler.DevSystemAdminHandler) {
	devAuth := middleware.RequireDevAPIKey(cfg.Server.DevAPIKey)

	if cfg.Server.EnablePprof {
		debug := r.Group("/debug")
		debug.Use(devAuth)
		pprof.Register(debug)
	}

	api := r.Group("/api")
	api.Use(devAuth)
	{
		api.GET("/status", handler.Status)

		system := api.Group("/system/admin")
		{
			system.POST("", devSystemAdmin.Create)
			system.GET("", devSystemAdmin.GetAll)
			system.GET("/:id", devSystemAdmin.GetByID)
			system.DELETE("/:id", devSystemAdmin.Delete)
		}
	}
}
