// Package apierr provides structured API error responses with machine-readable
// error codes that the frontend can use to display custom, localised messages.
//
// Every response that is not 2xx follows this JSON shape:
//
//	{
//	  "error":   "err_invalid_credentials",
//	  "message": "The email or password is incorrect."
//	}
//
// "error" is the stable, machine-readable code documented in docs/error-codes.md.
// "message" is a human-readable English fallback; frontends should prefer their
// own localised copy keyed by "error".
package apierr

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Code is a stable, machine-readable error identifier.
type Code string

const (
	// ── Request / input ─────────────────────────────────────────────────────
	CodeInvalidTenantID Code = "err_invalid_tenant_id"
	CodeInvalidID       Code = "err_invalid_id"
	CodeInvalidRequest  Code = "err_invalid_request"

	// ── Authentication ───────────────────────────────────────────────────────
	CodeMissingAuth         Code = "err_missing_auth"
	CodeInvalidCredentials  Code = "err_invalid_credentials"
	CodeAccountLocked       Code = "err_account_locked"
	CodeTokenInvalid        Code = "err_token_invalid"
	CodeTokenExpired        Code = "err_token_expired"
	CodeTokenAlreadyUsed    Code = "err_token_used"
	CodeInvalidTokenClaims  Code = "err_invalid_token_claims"

	// ── Authorisation ────────────────────────────────────────────────────────
	CodeForbidden Code = "err_forbidden"

	// ── Email verification ───────────────────────────────────────────────────
	CodeEmailNotVerified         Code = "err_email_not_verified"
	CodeVerificationTokenInvalid Code = "err_verification_token_invalid"

	// ── Resources ────────────────────────────────────────────────────────────
	CodeNotFound       Code = "err_not_found"
	CodeDuplicateEmail Code = "err_duplicate_email"

	// ── Internal ─────────────────────────────────────────────────────────────
	CodeInternal Code = "err_internal"
)

// defaultMessages provides English fallback text for each code.
// Frontends should key off the "error" field and supply their own copy.
var defaultMessages = map[Code]string{
	CodeInvalidTenantID:    "The provided tenant ID is not a valid UUID.",
	CodeInvalidID:          "The provided ID is not a valid UUID.",
	CodeInvalidRequest:     "The request body is missing or contains invalid JSON.",
	CodeMissingAuth:        "Authorization header is missing. Provide a Bearer token.",
	CodeInvalidCredentials: "The email or password is incorrect.",
	CodeAccountLocked:      "This account is temporarily locked due to too many failed login attempts. Please try again later.",
	CodeTokenInvalid:       "The token is invalid or has been tampered with.",
	CodeTokenExpired:       "The token has expired. Please log in again.",
	CodeTokenAlreadyUsed:   "This refresh token has already been used. Please log in again.",
	CodeInvalidTokenClaims: "The token contains invalid claims.",
	CodeForbidden:                "You do not have permission to perform this action.",
	CodeEmailNotVerified:         "Your email address has not been verified. Please check your inbox.",
	CodeVerificationTokenInvalid: "The verification token is invalid or has expired.",
	CodeNotFound:                 "The requested resource was not found.",
	CodeDuplicateEmail:     "An account with this email address already exists.",
	CodeInternal:           "An unexpected error occurred. Please try again later.",
}

type response struct {
	Error   Code   `json:"error"`
	Message string `json:"message"`
}

// Abort writes a structured error response and aborts the request chain.
func Abort(c *gin.Context, status int, code Code) {
	c.AbortWithStatusJSON(status, response{
		Error:   code,
		Message: defaultMessages[code],
	})
}

// AbortMsg is like Abort but overrides the English message (use sparingly —
// prefer a dedicated Code so the frontend can key off it).
func AbortMsg(c *gin.Context, status int, code Code, msg string) {
	c.AbortWithStatusJSON(status, response{Error: code, Message: msg})
}

// JSON writes a structured error response without aborting.
func JSON(c *gin.Context, status int, code Code) {
	c.JSON(status, response{
		Error:   code,
		Message: defaultMessages[code],
	})
}

// JSONMsg is like JSON but overrides the English message.
func JSONMsg(c *gin.Context, status int, code Code, msg string) {
	c.JSON(status, response{Error: code, Message: msg})
}

// JSONLocked writes a 429 account-locked response and sets the standard
// Retry-After header so clients know how long to wait before retrying.
func JSONLocked(c *gin.Context, retryAfter time.Duration) {
	secs := int(retryAfter.Seconds())
	if secs < 1 {
		secs = 1
	}
	c.Header("Retry-After", strconv.Itoa(secs))
	c.JSON(http.StatusTooManyRequests, response{
		Error:   CodeAccountLocked,
		Message: defaultMessages[CodeAccountLocked],
	})
}

// AbortLocked is like JSONLocked but also aborts the request chain.
func AbortLocked(c *gin.Context, retryAfter time.Duration) {
	secs := int(retryAfter.Seconds())
	if secs < 1 {
		secs = 1
	}
	c.Header("Retry-After", strconv.Itoa(secs))
	c.AbortWithStatusJSON(http.StatusTooManyRequests, response{
		Error:   CodeAccountLocked,
		Message: defaultMessages[CodeAccountLocked],
	})
}
