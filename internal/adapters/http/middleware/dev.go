package middleware

import (
	"net/http"

	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/apierr"
	"github.com/gin-gonic/gin"
)

// RequireDevAPIKey enforces that requests carry the correct X-Dev-API-Key header.
func RequireDevAPIKey(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Dev-API-Key") != key {
			apierr.AbortMsg(c, http.StatusUnauthorized, apierr.CodeMissingAuth, "invalid or missing dev API key")
			return
		}
		c.Next()
	}
}
