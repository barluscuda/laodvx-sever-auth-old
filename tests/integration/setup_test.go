// Package integration contains end-to-end HTTP API tests that run against a
// real Postgres + Redis backend.  The test binary connects to whichever
// instances the environment variables below point to (see TestMain).
//
// Run locally (requires Postgres + Redis):
//
//	TEST_DB_HOST=localhost TEST_REDIS_HOST=localhost \
//	  go test -v -count=1 ./tests/integration/...
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/barluscuda/laodvx-server-auth/config"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/database"
	emailadapter "github.com/barluscuda/laodvx-server-auth/internal/adapters/email"
	httphandler "github.com/barluscuda/laodvx-server-auth/internal/adapters/http/handler"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/middleware"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/http/router"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/redis"
	"github.com/barluscuda/laodvx-server-auth/internal/adapters/repository"
	"github.com/barluscuda/laodvx-server-auth/internal/domain/model"
	svc "github.com/barluscuda/laodvx-server-auth/internal/domain/service"
	"github.com/barluscuda/laodvx-server-auth/internal/server"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ── shared state ────────────────────────────────────────────────────────────

var (
	testHTTP        *httptest.Server         // main API server
	testTenantID    uuid.UUID                // pre-seeded tenant record
	testSystemAdmin = struct{ username, password string }{"sysadmin", "Test@12345"}
	testDevAPIKey   = "ci-test-dev-key"
	testEmailSender *emailadapter.CaptureSender // captures verification tokens
)

// ── TestMain ─────────────────────────────────────────────────────────────────

func TestMain(m *testing.M) {
	// 1. Change to a temp dir so that EnsureSeasonKeySet writes keys there.
	tmp, err := os.MkdirTemp("", "laodvx-auth-ci-*")
	must(err)
	must(os.Chdir(tmp))
	defer os.RemoveAll(tmp)

	// 2. Configure the server via environment before config.Get() fires.
	setenv("SERVER_MODE", "test")
	setenv("DEV_API_KEY", testDevAPIKey)
	setenv("SERVER_IP", "127.0.0.1")
	setenv("SERVER_DOMAIN", "auth.localhost")
	setenv("SERVER_PORT", "0") // unused — we use httptest
	setenv("SERVER_DEV_IP", "127.0.0.1")
	setenv("SERVER_DEV_DOMAIN", "auth.localhost")
	setenv("SERVER_DEV_PORT", "0")

	setenv("DB_HOST", getenv("TEST_DB_HOST", "localhost"))
	setenv("DB_PORT", getenv("TEST_DB_PORT", "5432"))
	setenv("DB_USER", getenv("TEST_DB_USER", "postgres"))
	setenv("DB_PASSWORD", getenv("TEST_DB_PASSWORD", "postgres"))
	setenv("DB_NAME", getenv("TEST_DB_NAME", "dx_auth_test"))
	setenv("DB_SSLMODE", "disable")

	setenv("REDIS_HOST", getenv("TEST_REDIS_HOST", "localhost"))
	setenv("REDIS_PORT", getenv("TEST_REDIS_PORT", "6379"))
	setenv("REDIS_PASSWORD", getenv("TEST_REDIS_PASSWORD", ""))
	setenv("REDIS_CACHE_TTL", "5") // 5-second cache in tests

	setenv("JWT_ACCESS_KEYS_DIR", "access")
	setenv("JWT_REFRESH_KEYS_DIR", "refresh")
	setenv("JWT_ACCESS_EXPIRY", "60")    // 1 minute
	setenv("JWT_REFRESH_EXPIRY", "3600") // 1 hour

	// 3. Boot config (once).
	cfg := config.Get()

	gin.SetMode(gin.TestMode)

	// 4. Connect to real DB and Redis.
	db := database.Connect(cfg.Database)
	db.AutoMigrate(
		&model.Tenant{},
		&model.TenantUser{},
		&model.SystemAdmin{},
		&model.RefreshToken{},
	)

	rdb := redis.Connect(cfg.Redis)

	// 5. Wire repos, services, handlers — exactly like main.go.
	testEmailSender = emailadapter.NewCaptureSender()
	repos := repository.New(db, rdb, cfg.Redis.CacheTTL, cfg.RateLimit)
	svcs := svc.NewServices(repos, cfg, testEmailSender)

	handlers := server.Handlers{
		TenantUser:           httphandler.NewTenantUserHandler(svcs.TenantUser),
		Auth:           httphandler.NewAuthHandler(svcs.UserAuth),
		SystemAdmin:    httphandler.NewSystemAdminHandler(svcs.Tenant, svcs.SystemAdmin),
		SystemAuth:     httphandler.NewSystemAuthHandler(svcs.SystemAuth),
		TenantAdmin:      httphandler.NewTenantAdminHandler(svcs.TenantAdmin),
		DevSystemAdmin: httphandler.NewDevSystemAdminHandler(svcs.DevSystemAdmin),
	}

	// 6. Build a plain gin.Engine (no http.Server — httptest manages listening).
	app := gin.New()
	app.Use(gin.Recovery())

	// Extract tenant_id from X-Tenant-Id header (set by nginx)
	app.Use(middleware.ExtractTenantIDFromHeader())
	// Enforce header/path scope rules for tenant and system routes
	app.Use(middleware.EnforceRouteScope())
	// Validate tenant exists if tenant_id header is present (returns 404 if not found)
	app.Use(middleware.TenantNotFound(repos.Tenant))

	app.GET("/.well-known/jwks.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, cfg.JWT.AccessKeys.PublicJWKS())
	})

	system := app.Group("/system/api")
	router.SetupSystemAdminRouter(system, cfg, handlers.SystemAdmin, handlers.SystemAuth)

	api := app.Group("/api")
	router.SetupAPIRouter(api, cfg, handlers.TenantUser, handlers.Auth)

	userAdmin := app.Group("/api/admin")
	router.SetupTenantAdminRouter(userAdmin, cfg, handlers.TenantAdmin)

	testHTTP = httptest.NewServer(app)

	// 7. Seed: create a system-admin and a tenant for tests.
	seedSystemAdmin(svcs, testSystemAdmin.username, testSystemAdmin.password)
	testTenantID = seedTenant(svcs)

	// 8. Run tests.
	code := m.Run()

	testHTTP.Close()
	os.Exit(code)
}

// ── seed helpers ─────────────────────────────────────────────────────────────

func seedSystemAdmin(svcs *svc.Services, username, password string) {
	// DevSystemAdminService.Create is idempotent enough — if it fails (duplicate)
	// we just carry on; the record already exists.
	svcs.DevSystemAdmin.Create(username, password) //nolint:errcheck
}

func seedTenant(svcs *svc.Services) uuid.UUID {
	t, err := svcs.Tenant.Create(fmt.Sprintf("ci-test-tenant-%d", time.Now().UnixNano()))
	must(err)
	return t.UUID
}

// ── HTTP client helpers ──────────────────────────────────────────────────────

// apiURL builds a URL on the test HTTP server.
func apiURL(path string) string {
	return testHTTP.URL + path
}

// tenantPath builds /api/<rest>.
func tenantPath(rest string) string {
	return "/api" + rest
}

// withTenantHeader sets the X-Tenant-Id header for tenant-scoped requests.
func withTenantHeader() func(*http.Request) {
	return func(r *http.Request) {
		r.Header.Set(middleware.TenantIDHeader, testTenantID.String())
	}
}

// do sends a JSON request with the tenant header set.
func do(t *testing.T, method, path string, body any, headers ...func(*http.Request)) *http.Response {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		must(json.NewEncoder(&buf).Encode(body))
	}

	req, err := http.NewRequest(method, apiURL(path), &buf)
	must(err)
	// Set X-Tenant-Id header for tenant_id extraction
	req.Header.Set(middleware.TenantIDHeader, testTenantID.String())
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, h := range headers {
		h(req)
	}

	resp, err := http.DefaultClient.Do(req)
	must(err)
	return resp
}

// doSystem sends a JSON request without the tenant header (for system admin endpoints).
func doSystem(t *testing.T, method, path string, body any, headers ...func(*http.Request)) *http.Response {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		must(json.NewEncoder(&buf).Encode(body))
	}

	req, err := http.NewRequest(method, apiURL(path), &buf)
	must(err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, h := range headers {
		h(req)
	}

	resp, err := http.DefaultClient.Do(req)
	must(err)
	return resp
}

// withBearer adds an Authorization: Bearer header.
func withBearer(token string) func(*http.Request) {
	return func(r *http.Request) {
		r.Header.Set("Authorization", "Bearer "+token)
	}
}

// decodeBody decodes the response JSON into dst and returns the status code.
func decodeBody(t *testing.T, resp *http.Response, dst any) int {
	t.Helper()
	defer resp.Body.Close()
	must(json.NewDecoder(resp.Body).Decode(dst))
	return resp.StatusCode
}

// ── assertion helpers ────────────────────────────────────────────────────────

// errBody is the structured API error response shape.
type errBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// tokenBody is the successful token pair response.
type tokenBody struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// assertError asserts that the response has the given HTTP status and
// error code, and that the message field is non-empty.
func assertError(t *testing.T, resp *http.Response, wantStatus int, wantCode string) {
	t.Helper()
	var body errBody
	gotStatus := decodeBody(t, resp, &body)
	if gotStatus != wantStatus {
		t.Errorf("status: got %d, want %d", gotStatus, wantStatus)
	}
	if body.Error != wantCode {
		t.Errorf("error code: got %q, want %q", body.Error, wantCode)
	}
	if body.Message == "" {
		t.Error("message field must not be empty")
	}
}

// assertTokens asserts that the response is 200/201 with non-empty tokens.
func assertTokens(t *testing.T, resp *http.Response, wantStatus int) tokenBody {
	t.Helper()
	var body tokenBody
	gotStatus := decodeBody(t, resp, &body)
	if gotStatus != wantStatus {
		t.Errorf("status: got %d, want %d", gotStatus, wantStatus)
	}
	if body.AccessToken == "" {
		t.Error("access_token must not be empty")
	}
	if body.RefreshToken == "" {
		t.Error("refresh_token must not be empty")
	}
	return body
}

// ── misc ─────────────────────────────────────────────────────────────────────

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func setenv(key, val string) {
	if err := os.Setenv(key, val); err != nil {
		panic(err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
