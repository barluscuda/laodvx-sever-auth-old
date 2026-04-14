package handler

import (
	"net/http"

	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/apierr"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/dto"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/middleware"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/gin-gonic/gin"
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
	tenantID := middleware.MustTenantUUID(c)

	if email := c.Query("email"); email != "" {
		u, err := h.svc.GetByEmail(tenantID, email)
		if err != nil {
			apierr.FromService(c, err)
			return
		}
		c.JSON(http.StatusOK, dto.ToTenantUserResponse(u))
		return
	}

	users, err := h.svc.GetAll(tenantID)
	if err != nil {
		apierr.FromService(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToTenantUserResponseList(users))
}

// GET /api/admin/user/:id
func (h *TenantAdminHandler) GetByID(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	u, err := h.svc.GetByID(middleware.MustTenantUUID(c), id)
	if err != nil {
		apierr.FromService(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToTenantUserResponse(u))
}
