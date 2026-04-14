package handler

import (
	"errors"
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
	tenantID, ok := middleware.GetTenantUUID(c)
	if !ok {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidTenantID)
		return
	}

	var req dto.CreateTenantUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidRequest)
		return
	}

	if err := h.svc.Create(tenantID, req.Email, req.Password); err != nil {
		if errors.Is(err, ports.ErrDuplicateEmail) {
			apierr.JSON(c, http.StatusConflict, apierr.CodeDuplicateEmail)
		} else {
			apierr.JSON(c, http.StatusInternalServerError, apierr.CodeInternal)
		}
		return
	}

	c.Status(http.StatusAccepted)
}

// POST /api/auth/verify-email  — public
func (h *TenantUserHandler) VerifyEmail(c *gin.Context) {
	tenantID, ok := middleware.GetTenantUUID(c)
	if !ok {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidTenantID)
		return
	}

	var req dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidRequest)
		return
	}

	if err := h.svc.VerifyEmail(tenantID, req.Email, req.OTP); err != nil {
		if errors.Is(err, ports.ErrVerificationTokenInvalid) {
			apierr.JSON(c, http.StatusUnprocessableEntity, apierr.CodeVerificationTokenInvalid)
		} else {
			apierr.JSON(c, http.StatusInternalServerError, apierr.CodeInternal)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// POST /api/user/resend-verification  — public
func (h *TenantUserHandler) ResendVerification(c *gin.Context) {
	tenantID, ok := middleware.GetTenantUUID(c)
	if !ok {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidTenantID)
		return
	}

	var req dto.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidRequest)
		return
	}

	// Ignore error intentionally — always return 202 to avoid email enumeration.
	h.svc.ResendVerification(tenantID, req.Email) //nolint:errcheck
	c.Status(http.StatusAccepted)
}

// GET /api/user/me  — requires JWT (user or user_admin)
func (h *TenantUserHandler) GetMe(c *gin.Context) {
	tenantID, ok := middleware.GetTenantUUID(c)
	if !ok {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidTenantID)
		return
	}

	claims := c.MustGet(middleware.ClaimsKey).(*token.Claims)
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		apierr.JSON(c, http.StatusUnauthorized, apierr.CodeInvalidTokenClaims)
		return
	}

	u, err := h.svc.GetByID(tenantID, userID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			apierr.JSON(c, http.StatusNotFound, apierr.CodeNotFound)
		} else {
			apierr.JSON(c, http.StatusInternalServerError, apierr.CodeInternal)
		}
		return
	}

	c.JSON(http.StatusOK, dto.ToTenantUserResponse(u))
}
