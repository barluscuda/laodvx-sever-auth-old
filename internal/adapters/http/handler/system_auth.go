package handler

import (
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
	if !bindJSON(c, &req) {
		return
	}
	pair, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		apierr.FromService(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToTokenResponse(pair))
}

// POST /system/api/auth/refresh
func (h *SystemAuthHandler) Refresh(c *gin.Context) {
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
