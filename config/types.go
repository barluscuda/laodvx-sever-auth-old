package config

import (
	"time"

	"github.com/barluscuda/laodvx-server-auth/internal/pkg/token"
)

// Config is the top-level configuration object assembled from environment
// variables at startup via config.Get().
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	RateLimit RateLimitConfig
	Email     EmailConfig
}

// ServerConfig holds HTTP listener settings for the main and dev servers.
type ServerConfig struct {
	// SERVER_MODE — Gin run mode. "release" disables debug output. Default: release.
	Mode string

	// SERVER_IP — Bind IP for the main API server. Default: 127.0.0.1.
	IP string

	// SERVER_PORT — Listening port for the main API server. Default: 3220.
	Port string

	// SERVER_DEV_IP — Bind IP for the dev/internal server. Default: 127.0.0.1.
	DevIP string

	// SERVER_DEV_PORT — Listening port for the dev server. Default: 3221.
	DevPort string

	// ENABLE_PPROF — Expose pprof endpoints on the dev server (debug mode only). Default: false.
	EnablePprof bool

	// DEV_API_KEY — API key required for all dev server endpoints.
	// Must be set when SERVER_MODE is not "release".
	DevAPIKey string
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	// DB_HOST — PostgreSQL host. Default: localhost.
	Host string

	// DB_PORT — PostgreSQL port. Default: 5432.
	Port string

	// DB_USER — PostgreSQL username. Default: postgres.
	TenantUser string

	// DB_PASSWORD — PostgreSQL password. Default: postgres.
	Password string

	// DB_NAME — Database name. Default: dx_auth.
	Name string

	// DB_SSLMODE — PostgreSQL SSL mode (require | verify-full | disable). Default: require.
	SSLMode string

	// DB_MAX_OPEN_CONNS — Maximum open connections in the pool. Default: 25.
	MaxOpenConns int

	// DB_MAX_IDLE_CONNS — Maximum idle connections in the pool. Default: 25.
	MaxIdleConns int

	// DB_CONN_MAX_LIFETIME_SECONDS — Max lifetime of a connection (seconds). Default: 3600.
	ConnMaxLifetime time.Duration

	// DB_CONN_MAX_IDLE_TIME_SECONDS — Max idle time before a connection is closed (seconds). Default: 900.
	ConnMaxIdleTime time.Duration
}

func (d DatabaseConfig) DSN() string {
	return "host=" + d.Host +
		" port=" + d.Port +
		" user=" + d.TenantUser +
		" password=" + d.Password +
		" dbname=" + d.Name +
		" sslmode=" + d.SSLMode
}

// RedisConfig holds Redis connection and caching settings.
type RedisConfig struct {
	// REDIS_HOST — Redis host. Default: localhost.
	Host string

	// REDIS_PORT — Redis port. Default: 6379.
	Port string

	// REDIS_PASSWORD — Redis password (leave empty for no auth). Default: "".
	Password string

	// Redis DB index (not configurable via env; always 0).
	DB int

	// REDIS_CACHE_TTL — TTL for cached read-through entries (seconds). Default: 3600.
	CacheTTL time.Duration
}

// JWTConfig holds signing key sets and expiry settings for access and refresh tokens.
type JWTConfig struct {
	// JWT_ACCESS_KEYS_DIR — Subfolder name inside keys/ for access-token EC key pairs.
	// A P-256 key pair is auto-generated per season (YYYY-sN); expired seasons are
	// pruned on startup. Public keys are served at /.well-known/jwks.json.
	// Default: access.
	AccessKeys *token.KeySet

	// JWT_REFRESH_KEYS_DIR — Same auto-rotation behaviour as JWT_ACCESS_KEYS_DIR
	// but for refresh token signing keys. Default: refresh.
	RefreshKeys *token.KeySet

	// JWT_ACCESS_EXPIRY — Access token lifetime (seconds). Default: 900 (15 min).
	AccessExpiry time.Duration

	// JWT_REFRESH_EXPIRY — Refresh token lifetime (seconds). Default: 2592000 (30 days).
	RefreshExpiry time.Duration
}

// RateLimitConfig controls brute-force protection on the login endpoint.
type RateLimitConfig struct {
	// LOGIN_MAX_ATTEMPTS — Number of failed login attempts before the account is
	// temporarily locked. Default: 5.
	LoginMaxAttempts int

	// LOGIN_LOCKOUT_SECONDS — Duration of the lockout window (seconds). Default: 900 (15 min).
	LoginLockoutTTL time.Duration
}

// EmailConfig holds SMTP relay settings and OTP behaviour.
type EmailConfig struct {
	// EMAIL_SMTP_HOST — SMTP relay hostname. Leave empty to disable email sending (noop mode).
	SMTPHost string

	// EMAIL_SMTP_PORT — SMTP relay port. Default: 587.
	SMTPPort string

	// EMAIL_SMTP_USERNAME — SMTP auth username. Leave empty for unauthenticated relay.
	SMTPUsername string

	// EMAIL_SMTP_PASSWORD — SMTP auth password.
	SMTPPassword string

	// EMAIL_FROM_ADDRESS — Sender address for all transactional emails.
	// Default: no-reply@example.com.
	FromAddress string

	// EMAIL_OTP_EXPIRY_SECONDS — How long a registration OTP remains valid (seconds).
	// Default: 600 (10 min).
	OTPExpiry time.Duration
}
