package handler

import (
	"net/http"

	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/apierr"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/dto"
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/gin-gonic/gin"
)

type SystemAdminHandler struct {
	tenantSvc ports.TenantService
	userSvc   ports.SystemAdminService
}

func NewSystemAdminHandler(tenantSvc ports.TenantService, userSvc ports.SystemAdminService) *SystemAdminHandler {
	return &SystemAdminHandler{tenantSvc: tenantSvc, userSvc: userSvc}
}

// POST /system/api/tenant
func (h *SystemAdminHandler) Create(c *gin.Context) {
	var req dto.CreateTenantRequest
	if !bindJSON(c, &req) {
		return
	}
	t, err := h.tenantSvc.Create(req.TenantName, req.Label, req.Plan)
	if err != nil {
		apierr.FromService(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.ToTenantResponse(t))
}

// GET /system/api/tenant
func (h *SystemAdminHandler) GetAll(c *gin.Context) {
	tenants, err := h.tenantSvc.GetAll()
	if err != nil {
		apierr.FromService(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToTenantResponseList(tenants))
}

// GET /system/api/tenant/:tenantname
func (h *SystemAdminHandler) GetByID(c *gin.Context) {
	t, err := h.tenantSvc.GetByTenantName(c.Param("tenantname"))
	if err != nil {
		apierr.FromService(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToTenantResponse(t))
}

// PUT /system/api/tenant/:tenantname
func (h *SystemAdminHandler) Update(c *gin.Context) {
	var req dto.UpdateTenantRequest
	if !bindJSON(c, &req) {
		return
	}
	t, err := h.tenantSvc.Update(c.Param("tenantname"), req.Label)
	if err != nil {
		apierr.FromService(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToTenantResponse(t))
}

// DELETE /system/api/tenant/:tenantname
func (h *SystemAdminHandler) Delete(c *gin.Context) {
	if err := h.tenantSvc.Delete(c.Param("tenantname")); err != nil {
		apierr.FromService(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// POST /system/api/tenant/:tenantname/admin
func (h *SystemAdminHandler) SetTenantAdmin(c *gin.Context) {
	var req dto.SetTenantAdminRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.userSvc.SetTenantAdmin(c.Param("tenantname"), req.UserID); err != nil {
		apierr.FromService(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// POST /system/api/tenant/:tenantname/ban
func (h *SystemAdminHandler) BanTenant(c *gin.Context) {
	if err := h.tenantSvc.Ban(c.Param("tenantname")); err != nil {
		apierr.FromService(c, err)
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

	switch {
	case tenantName != "" && email != "":
		u, err := h.userSvc.GetByTenantAndEmail(tenantName, email)
		h.respondOne(c, u, err)
	case tenantName != "":
		users, err := h.userSvc.GetAllByTenantName(tenantName)
		h.respondList(c, users, err)
	case email != "":
		users, err := h.userSvc.GetAllByEmail(email)
		h.respondList(c, users, err)
	default:
		users, err := h.userSvc.GetAll()
		h.respondList(c, users, err)
	}
}

// GET /system/api/user/:id
func (h *SystemAdminHandler) GetUserByID(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	u, err := h.userSvc.GetByID(id)
	if err != nil {
		apierr.FromService(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToTenantUserResponse(u))
}

func (h *SystemAdminHandler) respondOne(c *gin.Context, u *model.TenantUser, err error) {
	if err != nil {
		apierr.FromService(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToTenantUserResponse(u))
}

func (h *SystemAdminHandler) respondList(c *gin.Context, users []model.TenantUser, err error) {
	if err != nil {
		apierr.FromService(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToTenantUserResponseList(users))
}
