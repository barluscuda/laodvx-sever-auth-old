package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/barluscuda/laodvx-server-auth/config"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/apierr"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/handler"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/middleware"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/router"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg     config.Config
	mainSrv *http.Server
	devSrv  *http.Server
}

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
)

type Handlers struct {
	TenantUser           *handler.TenantUserHandler
	Auth           *handler.AuthHandler
	SystemAdmin    *handler.SystemAdminHandler
	SystemAuth     *handler.SystemAuthHandler
	TenantAdmin      *handler.TenantAdminHandler
	DevSystemAdmin *handler.DevSystemAdminHandler
}

func New(cfg config.Config, tenantRepo ports.TenantRepository, h Handlers) *Server {
	gin.SetMode(cfg.Server.Mode)

	app := gin.New()
	app.Use(gin.Recovery())
	if cfg.Server.Mode == gin.DebugMode {
		app.Use(gin.Logger())
	}

	// Extract tenant_id from X-Tenant-Id header (set by nginx)
	app.Use(middleware.ExtractTenantIDFromHeader())
	// Enforce header/path scope rules for tenant and system routes
	app.Use(middleware.EnforceRouteScope())
	// Validate tenant exists if tenant_id header is present (returns 404 if not found)
	app.Use(middleware.TenantNotFound(tenantRepo))

	app.GET("/.well-known/jwks.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, cfg.JWT.AccessKeys.PublicJWKS())
	})

	system := app.Group("/system/api")
	router.SetupSystemAdminRouter(system, cfg, h.SystemAdmin, h.SystemAuth)

	api := app.Group("/api")
	router.SetupAPIRouter(api, cfg, h.TenantUser, h.Auth)

	userAdmin := app.Group("/api/admin")
	router.SetupTenantAdminRouter(userAdmin, cfg, h.TenantAdmin)

	// 404 handler for undefined routes
	app.NoRoute(func(c *gin.Context) {
		apierr.AbortMsg(c, http.StatusNotFound, apierr.CodeNotFound, "404 api not found")
	})

	var devSrv *http.Server
	if cfg.Server.Mode != gin.ReleaseMode {
		devIP := cfg.Server.DevIP
		isLoopback := (devIP == "127.0.0.1" || devIP == "::1" || devIP == "localhost" ||
			strings.HasSuffix(devIP, ".localhost"))
		if !isLoopback {
			log.Printf("WARNING: dev server is bound to %s — unauthenticated system admin endpoints are exposed on a non-loopback address", devIP)
		}
		dev := gin.New()
		dev.Use(gin.Recovery())
		if cfg.Server.Mode == gin.DebugMode {
			dev.Use(gin.Logger())
		}
		router.SetupDevRouter(dev, cfg, h.DevSystemAdmin)
		// 404 handler for undefined routes in dev server
		dev.NoRoute(func(c *gin.Context) {
			apierr.AbortMsg(c, http.StatusNotFound, apierr.CodeNotFound, "404 api not found")
		})
		devSrv = &http.Server{
			Addr:              devIP + ":" + cfg.Server.DevPort,
			Handler:           dev,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		}
	}

	return &Server{
		cfg: cfg,
		mainSrv: &http.Server{
			Addr:              cfg.Server.IP + ":" + cfg.Server.Port,
			Handler:           app,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
		devSrv: devSrv,
	}
}

func (s *Server) Run() {
	go func() {
		log.Printf("server listening on %s", s.mainSrv.Addr)
		log.Printf("server is running in PID: %d", os.Getpid())
		if err := s.mainSrv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server error:", err)
		}
	}()

	if s.devSrv != nil {
		go func() {
			log.Println("dev server listening on", s.devSrv.Addr)
			if err := s.devSrv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				log.Println("dev server error:", err)
			}
		}()
	}

	s.waitForShutdown()
}

func (s *Server) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if s.devSrv != nil {
		if err := s.devSrv.Shutdown(ctx); err != nil {
			log.Println("dev server shutdown error:", err)
		}
	}
	if err := s.mainSrv.Shutdown(ctx); err != nil {
		log.Fatal("forced shutdown:", err)
	}

	log.Println("server stopped")
}
