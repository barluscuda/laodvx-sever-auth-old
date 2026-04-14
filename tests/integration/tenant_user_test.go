package integration

import (
	"net/http"
	"testing"

	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/middleware"
	"github.com/google/uuid"
)

// ── registration ──────────────────────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	resp := do(t, http.MethodPost, tenantPath("/user"), map[string]string{
		"email":    uniqueEmail(),
		"password": "Valid@1234",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("status: got %d, want 202", resp.StatusCode)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	email := uniqueEmail()

	// First registration.
	resp := do(t, http.MethodPost, tenantPath("/user"), map[string]string{
		"email":    email,
		"password": "Valid@1234",
	})
	resp.Body.Close()

	// Second registration with the same email.
	resp = do(t, http.MethodPost, tenantPath("/user"), map[string]string{
		"email":    email,
		"password": "Valid@1234",
	})
	assertError(t, resp, http.StatusConflict, "err_duplicate_email")
}

func TestRegister_MissingBody(t *testing.T) {
	resp := do(t, http.MethodPost, tenantPath("/user"), nil)
	assertError(t, resp, http.StatusBadRequest, "err_invalid_request")
}

// ── GET /user/me ──────────────────────────────────────────────────────────────

func TestGetMe_Success(t *testing.T) {
	email := uniqueEmail()
	tokens := registerAndLogin(t, email, "Valid@1234")

	resp := do(t, http.MethodGet, tenantPath("/user/me"), nil, withBearer(tokens.AccessToken))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want 200", resp.StatusCode)
	}
}

func TestGetMe_NoToken(t *testing.T) {
	resp := do(t, http.MethodGet, tenantPath("/user/me"), nil)
	assertError(t, resp, http.StatusUnauthorized, "err_missing_auth")
}

func TestGetMe_InvalidToken(t *testing.T) {
	resp := do(t, http.MethodGet, tenantPath("/user/me"), nil, withBearer("bad.token.here"))
	assertError(t, resp, http.StatusUnauthorized, "err_token_invalid")
}

func TestGetMe_WrongTenant(t *testing.T) {
	// Log in on testTenantID, then try to call /me with a different tenant_id header.
	email := uniqueEmail()
	tokens := registerAndLogin(t, email, "Valid@1234")

	differentTenantID := uuid.New()
	resp := do(t, http.MethodGet, "/api/user/me",
		nil, withBearer(tokens.AccessToken), func(r *http.Request) {
			// Override X-Tenant-Id with a different (non-existent) tenant UUID
			r.Header.Set(middleware.TenantIDHeader, differentTenantID.String())
		})
	// TenantNotFound returns 404 for unknown tenant; if it happened to exist the
	// JWT tenant_id mismatch would yield 403.
	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 or 404, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}
