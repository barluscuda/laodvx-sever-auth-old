package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// ── system admin login ────────────────────────────────────────────────────────

func TestSystemLogin_Success(t *testing.T) {
	resp := doSystem(t, http.MethodPost, "/system/api/auth/login", map[string]string{
		"username": testSystemAdmin.username,
		"password": testSystemAdmin.password,
	})
	assertTokens(t, resp, http.StatusOK)
}

func TestSystemLogin_WrongPassword(t *testing.T) {
	resp := doSystem(t, http.MethodPost, "/system/api/auth/login", map[string]string{
		"username": testSystemAdmin.username,
		"password": "wrongpassword",
	})
	assertError(t, resp, http.StatusUnauthorized, "err_invalid_credentials")
}

func TestSystemLogin_MissingBody(t *testing.T) {
	resp := doSystem(t, http.MethodPost, "/system/api/auth/login", nil)
	assertError(t, resp, http.StatusBadRequest, "err_invalid_request")
}

func TestSystemRoute_TenantHeaderRejected(t *testing.T) {
	resp := do(t, http.MethodPost, "/system/api/auth/login", map[string]string{
		"username": testSystemAdmin.username,
		"password": testSystemAdmin.password,
	})
	assertError(t, resp, http.StatusForbidden, "err_forbidden")
}

// ── system admin refresh ──────────────────────────────────────────────────────

func TestSystemRefresh_Success(t *testing.T) {
	resp := doSystem(t, http.MethodPost, "/system/api/auth/login", map[string]string{
		"username": testSystemAdmin.username,
		"password": testSystemAdmin.password,
	})
	tokens := assertTokens(t, resp, http.StatusOK)

	resp = doSystem(t, http.MethodPost, "/system/api/auth/refresh", map[string]string{
		"refresh_token": tokens.RefreshToken,
	})
	assertTokens(t, resp, http.StatusOK)
}

func TestSystemRefresh_TokenAlreadyUsed(t *testing.T) {
	resp := doSystem(t, http.MethodPost, "/system/api/auth/login", map[string]string{
		"username": testSystemAdmin.username,
		"password": testSystemAdmin.password,
	})
	tokens := assertTokens(t, resp, http.StatusOK)

	// Use once — succeeds.
	resp = doSystem(t, http.MethodPost, "/system/api/auth/refresh", map[string]string{
		"refresh_token": tokens.RefreshToken,
	})
	assertTokens(t, resp, http.StatusOK)

	// Use again — rejected.
	resp = doSystem(t, http.MethodPost, "/system/api/auth/refresh", map[string]string{
		"refresh_token": tokens.RefreshToken,
	})
	assertError(t, resp, http.StatusUnauthorized, "err_token_used")
}

// ── system admin – tenant CRUD ────────────────────────────────────────────────

// systemAdminTokens logs in as the system admin and returns an access token.
func systemAdminTokens(t *testing.T) tokenBody {
	t.Helper()
	resp := doSystem(t, http.MethodPost, "/system/api/auth/login", map[string]string{
		"username": testSystemAdmin.username,
		"password": testSystemAdmin.password,
	})
	return assertTokens(t, resp, http.StatusOK)
}

func TestTenant_CreateAndGet(t *testing.T) {
	tokens := systemAdminTokens(t)

	// Create.
	name := fmt.Sprintf("tenant-%d", time.Now().UnixNano())
	resp := doSystem(t, http.MethodPost, "/system/api/tenant", map[string]string{"name": name},
		withBearer(tokens.AccessToken))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create tenant: got %d, want 201", resp.StatusCode)
	}

	var created struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	decodeBody(t, resp, &created)

	if created.ID == "" {
		t.Error("id must not be empty")
	}
	if created.Name != name {
		t.Errorf("name: got %q, want %q", created.Name, name)
	}

	// Get by ID.
	resp = doSystem(t, http.MethodGet, "/system/api/tenant/"+created.ID, nil,
		withBearer(tokens.AccessToken))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get tenant: got %d, want 200", resp.StatusCode)
	}
}

func TestTenant_GetAll(t *testing.T) {
	tokens := systemAdminTokens(t)

	resp := doSystem(t, http.MethodGet, "/system/api/tenant", nil, withBearer(tokens.AccessToken))
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("get all tenants: got %d, want 200", resp.StatusCode)
	}
}

func TestTenant_NotFound(t *testing.T) {
	tokens := systemAdminTokens(t)

	resp := doSystem(t, http.MethodGet, "/system/api/tenant/00000000-0000-0000-0000-000000000099", nil,
		withBearer(tokens.AccessToken))
	assertError(t, resp, http.StatusNotFound, "err_not_found")
}

func TestTenant_Unauthorized(t *testing.T) {
	// Missing token.
	resp := doSystem(t, http.MethodGet, "/system/api/tenant", nil)
	assertError(t, resp, http.StatusUnauthorized, "err_missing_auth")
}

func TestTenant_DeleteNotFound(t *testing.T) {
	tokens := systemAdminTokens(t)

	resp := doSystem(t, http.MethodDelete, "/system/api/tenant/00000000-0000-0000-0000-000000000099", nil,
		withBearer(tokens.AccessToken))
	assertError(t, resp, http.StatusNotFound, "err_not_found")
}
