package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/middleware"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// uniqueEmail returns an email address guaranteed to be unique within a test run.
func uniqueEmail() string {
	return fmt.Sprintf("user+%d@ci.test", time.Now().UnixNano())
}

// register creates a new user account and returns 201. Fatally fails on error.
func register(t *testing.T, email, password string) {
	t.Helper()
	resp := do(t, http.MethodPost, tenantPath("/user"), map[string]string{
		"email":    email,
		"password": password,
	})
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("register: unexpected status %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// verifyEmail retrieves the captured token for the email and calls the
// verify-email endpoint. Fatally fails if no token was captured or the
// endpoint returns an unexpected status.
func verifyEmail(t *testing.T, email string) {
	t.Helper()
	token, ok := testEmailSender.LastToken(email)
	if !ok {
		t.Fatalf("verifyEmail: no token captured for %q", email)
	}
	resp := do(t, http.MethodPost, tenantPath("/auth/verify-email"), map[string]string{
		"email": email,
		"otp":   token,
	})
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("verifyEmail: unexpected status %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// registerAndLogin creates a fresh user, verifies the email, and logs in,
// returning the token pair. Fatally fails if any step fails.
func registerAndLogin(t *testing.T, email, password string) tokenBody {
	t.Helper()
	register(t, email, password)
	verifyEmail(t, email)

	resp := do(t, http.MethodPost, tenantPath("/auth/login"), map[string]string{
		"email":    email,
		"password": password,
	})
	return assertTokens(t, resp, http.StatusOK)
}

// ── login ─────────────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	email := uniqueEmail()
	tokens := registerAndLogin(t, email, "Valid@1234")

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("expected non-empty tokens")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	email := uniqueEmail()
	registerAndLogin(t, email, "Valid@1234")

	resp := do(t, http.MethodPost, tenantPath("/auth/login"), map[string]string{
		"email":    email,
		"password": "Wrong@9999",
	})
	assertError(t, resp, http.StatusUnauthorized, "err_invalid_credentials")
}

func TestLogin_UnknownEmail(t *testing.T) {
	resp := do(t, http.MethodPost, tenantPath("/auth/login"), map[string]string{
		"email":    "nobody@ci.test",
		"password": "Valid@1234",
	})
	// Timing-equalised path — still returns same error code.
	assertError(t, resp, http.StatusUnauthorized, "err_invalid_credentials")
}

func TestLogin_InvalidTenantID(t *testing.T) {
	resp := do(t, http.MethodPost, "/api/auth/login", map[string]string{
		"email":    "x@ci.test",
		"password": "Valid@1234",
	}, func(r *http.Request) {
		// Override X-Tenant-Id header with a non-UUID value
		r.Header.Set(middleware.TenantIDHeader, "not-a-uuid")
	})
	// ExtractTenantIDFromHeader middleware fires first and validates UUID.
	assertError(t, resp, http.StatusBadRequest, "err_invalid_tenant_id")
}

func TestTenantRoute_MissingTenantHeader(t *testing.T) {
	resp := doSystem(t, http.MethodPost, "/api/auth/login", map[string]string{
		"email":    "x@ci.test",
		"password": "Valid@1234",
	})
	assertError(t, resp, http.StatusBadRequest, "err_invalid_tenant_id")
}

func TestLogin_MissingBody(t *testing.T) {
	resp := do(t, http.MethodPost, tenantPath("/auth/login"), nil)
	assertError(t, resp, http.StatusBadRequest, "err_invalid_request")
}

func TestLogin_AccountLocked(t *testing.T) {
	email := uniqueEmail()
	registerAndLogin(t, email, "Valid@1234")

	// Exhaust the failed-attempt threshold (LOGIN_MAX_ATTEMPTS=5 in test setup).
	for i := 0; i < 5; i++ {
		resp := do(t, http.MethodPost, tenantPath("/auth/login"), map[string]string{
			"email":    email,
			"password": "Wrong@bad",
		})
		resp.Body.Close()
	}

	// Next attempt must be rejected with 429 + Retry-After header.
	resp := do(t, http.MethodPost, tenantPath("/auth/login"), map[string]string{
		"email":    email,
		"password": "Wrong@bad",
	})
	assertError(t, resp, http.StatusTooManyRequests, "err_account_locked")

	if resp.Header.Get("Retry-After") == "" {
		t.Error("expected Retry-After header on 429 response")
	}
}

// ── refresh ───────────────────────────────────────────────────────────────────

func TestRefresh_Success(t *testing.T) {
	email := uniqueEmail()
	first := registerAndLogin(t, email, "Valid@1234")

	resp := do(t, http.MethodPost, tenantPath("/auth/refresh"), map[string]string{
		"refresh_token": first.RefreshToken,
	})
	second := assertTokens(t, resp, http.StatusOK)

	if second.AccessToken == first.AccessToken {
		t.Error("new access token should differ from the original")
	}
}

func TestRefresh_TokenAlreadyUsed(t *testing.T) {
	email := uniqueEmail()
	tokens := registerAndLogin(t, email, "Valid@1234")

	// First refresh — succeeds.
	resp := do(t, http.MethodPost, tenantPath("/auth/refresh"), map[string]string{
		"refresh_token": tokens.RefreshToken,
	})
	assertTokens(t, resp, http.StatusOK)

	// Second use of the same refresh token — must be rejected.
	resp = do(t, http.MethodPost, tenantPath("/auth/refresh"), map[string]string{
		"refresh_token": tokens.RefreshToken,
	})
	assertError(t, resp, http.StatusUnauthorized, "err_token_used")
}

func TestRefresh_InvalidToken(t *testing.T) {
	resp := do(t, http.MethodPost, tenantPath("/auth/refresh"), map[string]string{
		"refresh_token": "this.is.not.a.valid.jwt",
	})
	assertError(t, resp, http.StatusUnauthorized, "err_token_invalid")
}

func TestRefresh_MissingBody(t *testing.T) {
	resp := do(t, http.MethodPost, tenantPath("/auth/refresh"), nil)
	assertError(t, resp, http.StatusBadRequest, "err_invalid_request")
}
