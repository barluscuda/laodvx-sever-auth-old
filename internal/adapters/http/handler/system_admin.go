package handler

import (
	"errors"
	"net/http"

	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/apierr"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/dto"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SystemAdminHandler struct {
	svc     ports.TenantService
	userSvc ports.SystemAdminService
}

func NewSystemAdminHandler(svc ports.TenantService, userSvc ports.SystemAdminService) *SystemAdminHandler {
	return &SystemAdminHandler{svc: svc, userSvc: userSvc}
}

// POST /system/api/tenant
func (h *SystemAdminHandler) Create(c *gin.Context) {
	var req dto.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidRequest)
		return
	}

	t, err := h.svc.Create(req.TenantName, req.Label, req.Plan)
	if err != nil {
		apierr.JSON(c, http.StatusInternalServerError, apierr.CodeInternal)
		return
	}

	c.JSON(http.StatusCreated, dto.ToTenantResponse(t))
}

// GET /system/api/tenant
func (h *SystemAdminHandler) GetAll(c *gin.Context) {
	tenants, err := h.svc.GetAll()
	if err != nil {
		apierr.JSON(c, http.StatusInternalServerError, apierr.CodeInternal)
		return
	}

	c.JSON(http.StatusOK, dto.ToTenantResponseList(tenants))
}

// GET /system/api/tenant/:tenantname
func (h *SystemAdminHandler) GetByID(c *gin.Context) {
	tenantName := c.Param("tenantname")

	t, err := h.svc.GetByTenantName(tenantName)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			apierr.JSON(c, http.StatusNotFound, apierr.CodeNotFound)
		} else {
			apierr.JSON(c, http.StatusInternalServerError, apierr.CodeInternal)
		}
		return
	}

	c.JSON(http.StatusOK, dto.ToTenantResponse(t))
}

// PUT /system/api/tenant/:tenantname
func (h *SystemAdminHandler) Update(c *gin.Context) {
	tenantName := c.Param("tenantname")

	var req dto.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidRequest)
		return
	}

	t, err := h.svc.Update(tenantName, req.Label)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			apierr.JSON(c, http.StatusNotFound, apierr.CodeNotFound)
		} else {
			apierr.JSON(c, http.StatusInternalServerError, apierr.CodeInternal)
		}
		return
	}

	c.JSON(http.StatusOK, dto.ToTenantResponse(t))
}

// DELETE /system/api/tenant/:tenantname
func (h *SystemAdminHandler) Delete(c *gin.Context) {
	tenantName := c.Param("tenantname")

	if err := h.svc.Delete(tenantName); err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			apierr.JSON(c, http.StatusNotFound, apierr.CodeNotFound)
		} else {
			apierr.JSON(c, http.StatusInternalServerError, apierr.CodeInternal)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// POST /system/api/tenant/:tenantname/admin
func (h *SystemAdminHandler) SetTenantAdmin(c *gin.Context) {
	tenantName := c.Param("tenantname")

	var req dto.SetTenantAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidRequest)
		return
	}

	if err := h.userSvc.SetTenantAdmin(tenantName, req.UserID); err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			apierr.JSON(c, http.StatusNotFound, apierr.CodeNotFound)
		} else {
			apierr.JSON(c, http.StatusInternalServerError, apierr.CodeInternal)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// POST /system/api/tenant/:tenantname/ban
func (h *SystemAdminHandler) BanTenant(c *gin.Context) {
	tenantName := c.Param("tenantname")

	if err := h.svc.Ban(tenantName); err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			apierr.JSON(c, http.StatusNotFound, apierr.CodeNotFound)
		} else {
			apierr.JSON(c, http.StatusInternalServerError, apierr.CodeInternal)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// GET /system/api/user
// GET /system/api/user?tenant_name=<name>
// GET /system/api/user?email=<email>
// GET /system/api/user?tenant_name=<name>&email=<email>
func (h *SystemAdminHandler) GetAllUsers(c *gin.Context) {
	tenantName := c.Query("tenant_name")
	email := c.Query("email")

	if tenantName != "" && email != "" {
		u, err := h.userSvc.GetByTenantAndEmail(tenantName, email)
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

	if tenantName != "" {
		users, err := h.userSvc.GetAllByTenantName(tenantName)
		if err != nil {
			apierr.JSON(c, http.StatusInternalServerError, apierr.CodeInternal)
			return
		}
		c.JSON(http.StatusOK, dto.ToTenantUserResponseList(users))
		return
	}

	if email != "" {
		users, err := h.userSvc.GetAllByEmail(email)
		if err != nil {
			apierr.JSON(c, http.StatusInternalServerError, apierr.CodeInternal)
			return
		}
		c.JSON(http.StatusOK, dto.ToTenantUserResponseList(users))
		return
	}

	users, err := h.userSvc.GetAll()
	if err != nil {
		apierr.JSON(c, http.StatusInternalServerError, apierr.CodeInternal)
		return
	}
	c.JSON(http.StatusOK, dto.ToTenantUserResponseList(users))
}

// GET /system/api/user/:id
func (h *SystemAdminHandler) GetUserByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidID)
		return
	}

	u, err := h.userSvc.GetByID(id)
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
