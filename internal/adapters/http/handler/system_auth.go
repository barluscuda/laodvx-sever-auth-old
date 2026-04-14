package handler

import (
	"errors"
	"net/http"

	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/apierr"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/dto"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/gin-gonic/gin"
)

type SystemAuthHandler struct {
	svc ports.SystemAuthService
}

func NewSystemAuthHandler(svc ports.SystemAuthService) *SystemAuthHandler {
	return &SystemAuthHandler{svc: svc}
}

// POST /system/api/auth/login
func (h *SystemAuthHandler) Login(c *gin.Context) {
	var req dto.SystemLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidRequest)
		return
	}

	pair, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		var lockErr *ports.AccountLockedError
		if errors.As(err, &lockErr) {
			apierr.JSONLocked(c, lockErr.RetryAfter)
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

// POST /system/api/auth/refresh
func (h *SystemAuthHandler) Refresh(c *gin.Context) {
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
