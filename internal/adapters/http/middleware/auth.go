package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/apierr"
	"github.com/barluscuda/laodvx-server-auth/internal/pkg/token"
	"github.com/barluscuda/laodvx-server-auth/internal/ports"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const ClaimsKey = "claims"

// RequireAuth validates the Bearer JWT (ES256) and checks that the caller has one of the allowed roles.
// For user/user_admin tokens it also verifies the token's tenant_id matches the context tenant_id (from header).
func RequireAuth(ks *token.KeySet, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			apierr.Abort(c, http.StatusUnauthorized, apierr.CodeMissingAuth)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &token.Claims{}
		t, err := jwt.ParseWithClaims(tokenStr, claims, ks.KeyFunc)
		if err != nil || !t.Valid {
			apierr.Abort(c, http.StatusUnauthorized, apierr.CodeTokenInvalid)
			return
		}

		// Role check
		if len(roles) > 0 && !slices.Contains(roles, claims.Role) {
			apierr.Abort(c, http.StatusForbidden, apierr.CodeForbidden)
			return
		}

		// For user/user_admin: token must belong to this tenant
		if claims.Role != ports.RoleSystemAdmin {
			if contextTenantID := GetTenantID(c); contextTenantID != "" {
				if claims.TenantID != contextTenantID {
					apierr.Abort(c, http.StatusForbidden, apierr.CodeForbidden)
					return
				}
			}
		}

		c.Set(ClaimsKey, claims)
		c.Next()
	}
}
