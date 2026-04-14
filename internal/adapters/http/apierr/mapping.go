package apierr

import (
	"errors"
	"net/http"

	"github.com/barluscuda/laodvx-server-auth/internal/ports"
	"github.com/gin-gonic/gin"
)

// FromService writes a structured response for a service-layer error,
// centralizing the mapping of domain errors to HTTP status + error code.
// Returns true if the error was handled (caller should return early).
func FromService(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	var lockErr *ports.AccountLockedError
	if errors.As(err, &lockErr) {
		JSONLocked(c, lockErr.RetryAfter)
		return true
	}

	status, code := classify(err)
	JSON(c, status, code)
	return true
}

func classify(err error) (int, Code) {
	switch {
	case errors.Is(err, ports.ErrNotFound):
		return http.StatusNotFound, CodeNotFound
	case errors.Is(err, ports.ErrDuplicateEmail):
		return http.StatusConflict, CodeDuplicateEmail
	case errors.Is(err, ports.ErrRegistrationPending):
		return http.StatusConflict, CodeRegistrationPending
	case errors.Is(err, ports.ErrOTPRateLimited):
		return http.StatusTooManyRequests, CodeOTPRateLimited
	case errors.Is(err, ports.ErrInvalidCredentials):
		return http.StatusUnauthorized, CodeInvalidCredentials
	case errors.Is(err, ports.ErrEmailNotVerified):
		return http.StatusForbidden, CodeEmailNotVerified
	case errors.Is(err, ports.ErrVerificationTokenInvalid):
		return http.StatusUnprocessableEntity, CodeVerificationTokenInvalid
	case errors.Is(err, ports.ErrTokenAlreadyUsed):
		return http.StatusUnauthorized, CodeTokenAlreadyUsed
	case errors.Is(err, ports.ErrTokenExpired):
		return http.StatusUnauthorized, CodeTokenExpired
	case errors.Is(err, ports.ErrInvalidToken):
		return http.StatusUnauthorized, CodeTokenInvalid
	default:
		return http.StatusInternalServerError, CodeInternal
	}
}
