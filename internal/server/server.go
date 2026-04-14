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

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
)

type Server struct {
	mainSrv *http.Server
	devSrv  *http.Server
}

type Handlers struct {
	TenantUser     *handler.TenantUserHandler
	Auth           *handler.AuthHandler
	SystemAdmin    *handler.SystemAdminHandler
	SystemAuth     *handler.SystemAuthHandler
	TenantAdmin    *handler.TenantAdminHandler
	DevSystemAdmin *handler.DevSystemAdminHandler
}

// New constructs the Server. It builds the main HTTP engine and, in non-release
// modes, an additional dev engine bound to cfg.Server.DevIP.
func New(cfg config.Config, tenantRepo ports.TenantRepository, h Handlers) *Server {
	gin.SetMode(cfg.Server.Mode)

	return &Server{
		mainSrv: newMainHTTPServer(cfg, tenantRepo, h),
		devSrv:  newDevHTTPServer(cfg, h),
	}
}

func newMainHTTPServer(cfg config.Config, tenantRepo ports.TenantRepository, h Handlers) *http.Server {
	app := gin.New()
	installBaseMiddleware(app, cfg)

	// Tenant routing: extract from header, enforce scope, validate existence.
	app.Use(middleware.ExtractTenantIDFromHeader())
	app.Use(middleware.EnforceRouteScope())
	app.Use(middleware.TenantNotFound(tenantRepo))

	app.GET("/.well-known/jwks.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, cfg.JWT.AccessKeys.PublicJWKS())
	})

	router.SetupSystemAdminRouter(app.Group("/system/api"), cfg, h.SystemAdmin, h.SystemAuth)
	router.SetupAPIRouter(app.Group("/api"), cfg, h.TenantUser, h.Auth)
	router.SetupTenantAdminRouter(app.Group("/api/admin"), cfg, h.TenantAdmin)

	app.NoRoute(notFoundHandler)

	return buildHTTPServer(cfg.Server.IP+":"+cfg.Server.Port, app)
}

func newDevHTTPServer(cfg config.Config, h Handlers) *http.Server {
	if cfg.Server.Mode == gin.ReleaseMode {
		return nil
	}

	if !isLoopbackAddr(cfg.Server.DevIP) {
		log.Printf("WARNING: dev server is bound to %s — unauthenticated system admin endpoints are exposed on a non-loopback address", cfg.Server.DevIP)
	}

	dev := gin.New()
	installBaseMiddleware(dev, cfg)
	router.SetupDevRouter(dev, cfg, h.DevSystemAdmin)
	dev.NoRoute(notFoundHandler)

	return buildHTTPServer(cfg.Server.DevIP+":"+cfg.Server.DevPort, dev)
}

func installBaseMiddleware(app *gin.Engine, cfg config.Config) {
	app.Use(gin.Recovery())
	if cfg.Server.Mode == gin.DebugMode {
		app.Use(gin.Logger())
	}
}

func buildHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

func notFoundHandler(c *gin.Context) {
	apierr.AbortMsg(c, http.StatusNotFound, apierr.CodeNotFound, "404 api not found")
}

func isLoopbackAddr(ip string) bool {
	return ip == "127.0.0.1" || ip == "::1" || ip == "localhost" || strings.HasSuffix(ip, ".localhost")
}

// Run starts the HTTP servers and blocks until SIGINT/SIGTERM is received.
func (s *Server) Run() {
	go s.serve(s.mainSrv, "server")
	if s.devSrv != nil {
		go s.serve(s.devSrv, "dev server")
	}
	s.waitForShutdown()
}

func (s *Server) serve(srv *http.Server, name string) {
	log.Printf("%s listening on %s (PID %d)", name, srv.Addr, os.Getpid())
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("%s error: %v", name, err)
	}
}

func (s *Server) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
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
