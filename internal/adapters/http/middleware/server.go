package middleware

import (
	"net/http"
	"path"
	"strings"

	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/apierr"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const TenantIDKey = "tenant_id"

// TenantIDHeader is the HTTP header nginx sets when reverse-proxying tenant requests.
const TenantIDHeader = "X-Tenant-Id"

type routeScope int

const (
	routeScopeNeutral routeScope = iota
	routeScopeTenantAPI
	routeScopeSystemAPI
)

// GetTenantUUID retrieves the parsed tenant UUID from the request context.
// Returns uuid.Nil, false when not found or stored in an unexpected format.
func GetTenantUUID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(TenantIDKey)
	if !exists {
		return uuid.Nil, false
	}
	tenantID, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, false
	}
	return tenantID, true
}

// GetTenantID retrieves the tenant_id string from the request context.
// Returns an empty string if not found.
func GetTenantID(c *gin.Context) string {
	if tenantID, ok := GetTenantUUID(c); ok {
		return tenantID.String()
	}
	return ""
}

// ExtractTenantIDFromHeader extracts the tenant_id from the X-Tenant-Id header
// set by nginx when reverse-proxying tenant requests.
// If the header is absent the context is left without a tenant_id (system routes).
// If the header is present but not a valid UUID the request is rejected with 400.
func ExtractTenantIDFromHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimSpace(c.GetHeader(TenantIDHeader))
		if raw == "" {
			c.Next()
			return
		}

		tenantID, err := uuid.Parse(raw)
		if err != nil {
			apierr.Abort(c, http.StatusBadRequest, apierr.CodeInvalidTenantID)
			return
		}

		c.Set(TenantIDKey, tenantID)
		c.Next()
	}
}

// EnforceRouteScope ensures that tenant API routes are called with a tenant_id header
// and system API routes are called without one.
func EnforceRouteScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, hasTenantID := GetTenantUUID(c)

		switch classifyRouteScope(c.Request.URL.Path) {
		case routeScopeSystemAPI:
			if hasTenantID {
				apierr.Abort(c, http.StatusForbidden, apierr.CodeForbidden)
				return
			}
		case routeScopeTenantAPI:
			if !hasTenantID {
				apierr.Abort(c, http.StatusBadRequest, apierr.CodeInvalidTenantID)
				return
			}
		}

		c.Next()
	}
}

func classifyRouteScope(requestPath string) routeScope {
	cleanPath := path.Clean("/" + strings.TrimSpace(requestPath))

	switch {
	case cleanPath == "/system/api" || strings.HasPrefix(cleanPath, "/system/api/"):
		return routeScopeSystemAPI
	case cleanPath == "/api" || strings.HasPrefix(cleanPath, "/api/"):
		return routeScopeTenantAPI
	default:
		return routeScopeNeutral
	}
}

// TenantNotFound validates that a tenant exists when a tenant_id header is present.
// If tenant_id is in the context but the tenant doesn't exist in the database, returns 404.
// If no tenant_id is present, skips this middleware.
func TenantNotFound(repo ports.TenantRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, ok := GetTenantUUID(c)
		if !ok {
			c.Next()
			return
		}

		exists, err := repo.ExistsByID(tenantID)
		if err != nil {
			apierr.Abort(c, http.StatusInternalServerError, apierr.CodeInternal)
			return
		}
		if !exists {
			apierr.Abort(c, http.StatusNotFound, apierr.CodeNotFound)
			return
		}

		c.Next()
	}
}
