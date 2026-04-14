package handler

import (
	"net/http"

	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/apierr"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/dto"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/gin-gonic/gin"
)

type DevSystemAdminHandler struct {
	svc ports.DevSystemAdminService
}

func NewDevSystemAdminHandler(svc ports.DevSystemAdminService) *DevSystemAdminHandler {
	return &DevSystemAdminHandler{svc: svc}
}

// POST /api/system-admin
func (h *DevSystemAdminHandler) Create(c *gin.Context) {
	var req dto.CreateSystemAdminRequest
	if !bindJSON(c, &req) {
		return
	}
	admin, err := h.svc.Create(req.Username, req.Password)
	if err != nil {
		apierr.FromService(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.ToSystemAdminResponse(admin))
}

// GET /api/system-admin
func (h *DevSystemAdminHandler) GetAll(c *gin.Context) {
	admins, err := h.svc.GetAll()
	if err != nil {
		apierr.FromService(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToSystemAdminResponseList(admins))
}

// GET /api/system-admin/:id
func (h *DevSystemAdminHandler) GetByID(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	admin, err := h.svc.GetByID(id)
	if err != nil {
		apierr.FromService(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToSystemAdminResponse(admin))
}

// DELETE /api/system-admin/:id
func (h *DevSystemAdminHandler) Delete(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(id); err != nil {
		apierr.FromService(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
