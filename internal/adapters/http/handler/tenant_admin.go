package handler

import (
	"errors"
	"net/http"

	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/apierr"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/dto"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/middleware"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TenantAdminHandler struct {
	svc ports.TenantAdminService
}

func NewTenantAdminHandler(svc ports.TenantAdminService) *TenantAdminHandler {
	return &TenantAdminHandler{svc: svc}
}

// GET /api/admin/user
// GET /api/admin/user?email=<email>
func (h *TenantAdminHandler) GetAll(c *gin.Context) {
	tenantID, ok := middleware.GetTenantUUID(c)
	if !ok {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidTenantID)
		return
	}

	if email := c.Query("email"); email != "" {
		u, err := h.svc.GetByEmail(tenantID, email)
		if err != nil {
			if errors.Is(err, ports.ErrNotFound) {
				apierr.JSON(c, http.StatusNotFound, apierr.CodeNotFound)
			} else {
				apierr.JSON(c, http.StatusInternalServerError, apierr.CodeInternal)
			}
			return
		}
		c.JSON(http.StatusOK, dto.ToTenantUserResponse(u))
		return
	}

	users, err := h.svc.GetAll(tenantID)
	if err != nil {
		apierr.JSON(c, http.StatusInternalServerError, apierr.CodeInternal)
		return
	}
	c.JSON(http.StatusOK, dto.ToTenantUserResponseList(users))
}

// GET /api/admin/user/:id
func (h *TenantAdminHandler) GetByID(c *gin.Context) {
	tenantID, ok := middleware.GetTenantUUID(c)
	if !ok {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidTenantID)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidID)
		return
	}

	u, err := h.svc.GetByID(tenantID, id)
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
