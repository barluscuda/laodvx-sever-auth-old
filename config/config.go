package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/barluscuda/laodvx-server-auth/internal/pkg/token"
	"github.com/joho/godotenv"
)

var (
	instance Config
	loadOnce sync.Once
)

func Load() {
	loadOnce.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Println("no .env file found, using only system environment variables")
		}

		devAPIKey := getEnv("DEV_API_KEY", "")
		if cfg := getEnv("SERVER_MODE", "release"); cfg != "release" && devAPIKey == "" {
			log.Fatal("DEV_API_KEY must be set when running in non-release mode — refusing to start")
		}

		instance = Config{
			Server: ServerConfig{
				Mode:        getEnv("SERVER_MODE", "release"),
				IP:          getEnv("SERVER_IP", "127.0.0.1"),
				Port:        getEnv("SERVER_PORT", "3220"),
				DevIP:       getEnv("SERVER_DEV_IP", "127.0.0.1"),
				DevPort:     getEnv("SERVER_DEV_PORT", "3221"),
				EnablePprof: getEnv("ENABLE_PPROF", "false") == "true",
				DevAPIKey:   devAPIKey,
			},
			Database: DatabaseConfig{
				Host:            getEnv("DB_HOST", "localhost"),
				Port:            getEnv("DB_PORT", "5432"),
				TenantUser:      getEnv("DB_USER", "postgres"),
				Password:        getEnv("DB_PASSWORD", "postgres"),
				Name:            getEnv("DB_NAME", "dx_auth"),
				SSLMode:         getEnv("DB_SSLMODE", "require"),
				MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
				MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 25),
				ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME_SECONDS", time.Hour),
				ConnMaxIdleTime: getDurationEnv("DB_CONN_MAX_IDLE_TIME_SECONDS", 15*time.Minute),
			},
			Redis: RedisConfig{
				Host:     getEnv("REDIS_HOST", "localhost"),
				Port:     getEnv("REDIS_PORT", "6379"),
				Password: getEnv("REDIS_PASSWORD", ""),
				DB:       0,
				CacheTTL: getDurationEnv("REDIS_CACHE_TTL", time.Hour),
			},
			JWT: JWTConfig{
				AccessKeys:    mustEnsureKeySet("JWT_ACCESS_KEYS_DIR"),
				RefreshKeys:   mustEnsureKeySet("JWT_REFRESH_KEYS_DIR"),
				AccessExpiry:  getDurationEnv("JWT_ACCESS_EXPIRY", 15*time.Minute),
				RefreshExpiry: getDurationEnv("JWT_REFRESH_EXPIRY", 30*24*time.Hour),
			},
			RateLimit: RateLimitConfig{
				LoginMaxAttempts: getIntEnv("LOGIN_MAX_ATTEMPTS", 5),
				LoginLockoutTTL:  getDurationEnv("LOGIN_LOCKOUT_SECONDS", 15*time.Minute),
			},
			Email: EmailConfig{
				SMTPHost:     getEnv("EMAIL_SMTP_HOST", ""),
				SMTPPort:     getEnv("EMAIL_SMTP_PORT", "587"),
				SMTPUsername: getEnv("EMAIL_SMTP_USERNAME", ""),
				SMTPPassword: getEnv("EMAIL_SMTP_PASSWORD", ""),
				FromAddress:  getEnv("EMAIL_FROM_ADDRESS", "no-reply@example.com"),
				OTPExpiry:            getDurationEnv("EMAIL_OTP_EXPIRY_SECONDS", 10*time.Minute),
				OTPMaxIssues:         getIntEnv("EMAIL_OTP_MAX_ISSUES", 5),
				OTPResendCooldown:    getDurationEnv("EMAIL_OTP_RESEND_COOLDOWN_SECONDS", 60*time.Second),
				OTPMaxVerifyAttempts: getIntEnv("EMAIL_OTP_MAX_VERIFY_ATTEMPTS", 5),
				OTPVerifyLockoutTTL:  getDurationEnv("EMAIL_OTP_VERIFY_LOCKOUT_SECONDS", 15*time.Minute),
			},
		}
	})
}

func Get() Config {
	Load()
	return instance
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("warning: %s=%q is not a valid integer, using default %d", key, v, fallback)
		return fallback
	}
	return n
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	secs, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("warning: %s=%q is not a valid integer (seconds), using default %s", key, v, fallback)
		return fallback
	}
	return time.Duration(secs) * time.Second
}

func mustEnsureKeySet(nameEnvKey string) *token.KeySet {
	name := getEnv(nameEnvKey, "")
	if name == "" {
		log.Fatalf("%s must be set — refusing to start", nameEnvKey)
	}
	dir := filepath.Join("keys", name)
	ks, err := token.EnsureSeasonKeySet(dir)
	if err != nil {
		log.Fatalf("failed to ensure key set (%s=%q): %v", nameEnvKey, dir, err)
	}
	return ks
}
