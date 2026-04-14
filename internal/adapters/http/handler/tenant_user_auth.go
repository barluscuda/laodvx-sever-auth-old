package handler

import (
	"errors"
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
	tenantID, ok := middleware.GetTenantUUID(c)
	if !ok {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidTenantID)
		return
	}

	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidRequest)
		return
	}

	pair, err := h.svc.Login(tenantID, req.Email, req.Password)
	if err != nil {
		var lockErr *ports.AccountLockedError
		if errors.As(err, &lockErr) {
			apierr.JSONLocked(c, lockErr.RetryAfter)
			return
		}
		if errors.Is(err, ports.ErrEmailNotVerified) {
			apierr.JSON(c, http.StatusForbidden, apierr.CodeEmailNotVerified)
			return
		}
		if !errors.Is(err, ports.ErrInvalidCredentials) {
			apierr.JSON(c, http.StatusInternalServerError, apierr.CodeInternal)
			return
		}
		apierr.JSON(c, http.StatusUnauthorized, apierr.CodeInvalidCredentials)
		return
	}

	c.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}

// POST /api/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidRequest)
		return
	}

	pair, err := h.svc.Refresh(req.RefreshToken)
	if err != nil {
		if errors.Is(err, ports.ErrTokenAlreadyUsed) {
			apierr.JSON(c, http.StatusUnauthorized, apierr.CodeTokenAlreadyUsed)
			return
		}
		apierr.JSON(c, http.StatusUnauthorized, apierr.CodeTokenInvalid)
		return
	}

	c.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}
