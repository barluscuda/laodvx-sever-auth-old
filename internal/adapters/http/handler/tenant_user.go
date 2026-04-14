package handler

import (
	"net/http"

	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/apierr"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/dto"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/middleware"
	"github.com/barluscuda/laodvx-server-auth/internal/pkg/token"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TenantUserHandler struct {
	svc ports.TenantUserService
}

func NewTenantUserHandler(svc ports.TenantUserService) *TenantUserHandler {
	return &TenantUserHandler{svc: svc}
}

// POST /api/user  — public registration
func (h *TenantUserHandler) Create(c *gin.Context) {
	var req dto.CreateTenantUserRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.Create(middleware.MustTenantUUID(c), req.Email, req.Password); err != nil {
		apierr.FromService(c, err)
		return
	}
	c.Status(http.StatusAccepted)
}

// POST /api/auth/verify-email  — public
func (h *TenantUserHandler) VerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.VerifyEmail(middleware.MustTenantUUID(c), req.Email, req.OTP); err != nil {
		apierr.FromService(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// POST /api/user/resend-verification  — public
func (h *TenantUserHandler) ResendVerification(c *gin.Context) {
	var req dto.ResendVerificationRequest
	if !bindJSON(c, &req) {
		return
	}
	// Ignore error intentionally — always return 202 to avoid email enumeration.
	_ = h.svc.ResendVerification(middleware.MustTenantUUID(c), req.Email)
	c.Status(http.StatusAccepted)
}

// GET /api/user/me  — requires JWT (user or user_admin)
func (h *TenantUserHandler) GetMe(c *gin.Context) {
	claims := c.MustGet(middleware.ClaimsKey).(*token.Claims)
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		apierr.JSON(c, http.StatusUnauthorized, apierr.CodeInvalidTokenClaims)
		return
	}
	u, err := h.svc.GetByID(middleware.MustTenantUUID(c), userID)
	if err != nil {
		apierr.FromService(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToTenantUserResponse(u))
}
