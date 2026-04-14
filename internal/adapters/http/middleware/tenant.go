package middleware

import (
	"net/http"

	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/apierr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequireTenant aborts the request with 400 when no tenant UUID is present in
// the context. Handlers mounted behind this middleware can use MustTenantUUID
// without further checks.
func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := GetTenantUUID(c); !ok {
			apierr.Abort(c, http.StatusBadRequest, apierr.CodeInvalidTenantID)
			return
		}
		c.Next()
	}
}

// MustTenantUUID returns the tenant UUID from context. Panics if absent — only
// use from handlers mounted behind RequireTenant (or an equivalent middleware).
func MustTenantUUID(c *gin.Context) uuid.UUID {
	id, ok := GetTenantUUID(c)
	if !ok {
		panic("middleware.MustTenantUUID: tenant UUID missing — route must be mounted behind RequireTenant")
	}
	return id
}
