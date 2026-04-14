package handler

import (
	"net/http"

	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/apierr"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/dto"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/middleware"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc ports.TenantUserAuthService
}

func NewAuthHandler(svc ports.TenantUserAuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !bindJSON(c, &req) {
		return
	}
	pair, err := h.svc.Login(middleware.MustTenantUUID(c), req.Email, req.Password)
	if err != nil {
		apierr.FromService(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToTokenResponse(pair))
}

// POST /api/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if !bindJSON(c, &req) {
		return
	}
	pair, err := h.svc.Refresh(req.RefreshToken)
	if err != nil {
		apierr.FromService(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToTokenResponse(pair))
}
