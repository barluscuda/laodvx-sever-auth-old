package handler

import (
	"net/http"

	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/apierr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// bindJSON decodes the request body into req, writing 400 on failure.
// Returns true on success.
func bindJSON(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidRequest)
		return false
	}
	return true
}

// parseUUIDParam parses a URL path parameter as a UUID, writing 400 on failure.
// Returns (uuid, true) on success.
func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		apierr.JSON(c, http.StatusBadRequest, apierr.CodeInvalidID)
		return uuid.Nil, false
	}
	return id, true
}
