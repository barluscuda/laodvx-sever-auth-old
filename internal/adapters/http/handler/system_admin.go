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

	t, err := h.svc.Create(req.Name)
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

// GET /system/api/tenant/:id
func (h *SystemAdminHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidID)
		return
	}

	t, err := h.svc.GetByID(id)
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

// PUT /system/api/tenant/:id
func (h *SystemAdminHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidID)
		return
	}

	var req dto.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidRequest)
		return
	}

	t, err := h.svc.Update(id, req.Name)
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

// DELETE /system/api/tenant/:id
func (h *SystemAdminHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidID)
		return
	}

	if err := h.svc.Delete(id); err != nil {
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
// GET /system/api/user?tenant_id=<uuid>
// GET /system/api/user?email=<email>
// GET /system/api/user?tenant_id=<uuid>&email=<email>
func (h *SystemAdminHandler) GetAllUsers(c *gin.Context) {
	rawTenantID := c.Query("tenant_id")
	email := c.Query("email")

	if rawTenantID != "" && email != "" {
		tenantID, err := uuid.Parse(rawTenantID)
		if err != nil {
			apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidTenantID)
			return
		}
		u, err := h.userSvc.GetByTenantAndEmail(tenantID, email)
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

	if rawTenantID != "" {
		tenantID, err := uuid.Parse(rawTenantID)
		if err != nil {
			apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidTenantID)
			return
		}
		users, err := h.userSvc.GetAllByTenantID(tenantID)
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
