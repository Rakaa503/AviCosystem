package config

import (
	"os"
	"strings"
	"testing"
)

func clearConfigEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"APP_NAME", "APP_ENV", "APP_PORT",
		"CORS_ALLOWED_ORIGINS",
		"RATE_LIMIT_REQUESTS", "RATE_LIMIT_WINDOW_SECONDS",
		"REQUEST_TIMEOUT_SECONDS",
		"TRUST_PROXY", "TRUSTED_PROXIES",
		"DATABASE_HOST", "DATABASE_PORT", "DATABASE_USER",
		"DATABASE_PASSWORD", "DATABASE_NAME", "DATABASE_SSLMODE",
		"DATABASE_MAX_OPEN_CONNS", "DATABASE_MAX_IDLE_CONNS",
		"DATABASE_CONN_MAX_LIFETIME_SECONDS",
		"DATABASE_CONN_MAX_IDLE_TIME_SECONDS",
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB",
		"REDIS_POOL_SIZE", "REDIS_MIN_IDLE_CONNS", "REDIS_MAX_RETRIES",
		"REDIS_DIAL_TIMEOUT_SECONDS", "REDIS_READ_TIMEOUT_SECONDS",
		"REDIS_WRITE_TIMEOUT_SECONDS",
	}

	for _, key := range keys {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("failed to unset %s: %v", key, err)
		}
	}
}

func validConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:                   "AviCosystem",
			Env:                    "development",
			Port:                   "8080",
			CORSAllowedOrigins:     []string{"http://localhost:5173"},
			RateLimitRequests:      100,
			RateLimitWindowSeconds: 60,
			RequestTimeoutSeconds:  30,
		},
		Database: DatabaseConfig{
			Host:                   "localhost",
			Port:                   "5432",
			User:                   "postgres",
			Name:                   "avicosystem",
			MaxOpenConns:           25,
			MaxIdleConns:           10,
			ConnMaxLifetimeSeconds: 1800,
			ConnMaxIdleTimeSeconds: 300,
		},
		Redis: RedisConfig{
			Host:         "localhost",
			Port:         "6379",
			DB:           0,
			PoolSize:     10,
			MinIdleConns: 2,
			MaxRetries:   3,
			DialTimeout:  5,
			ReadTimeout:  3,
			WriteTimeout: 3,
		},
	}
}

func TestConfigDefaultValues(t *testing.T) {
	clearConfigEnv(t)

	cfg := &Config{
		App:      loadAppConfig(),
		Database: loadDatabaseConfig(),
		Redis:    loadRedisConfig(),
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	if cfg.App.Name != "AviCosystem" {
		t.Errorf("unexpected app name: %s", cfg.App.Name)
	}

	if cfg.App.Port != "8080" {
		t.Errorf("unexpected app port: %s", cfg.App.Port)
	}

	if cfg.Database.Port != "5432" {
		t.Errorf("unexpected database port: %s", cfg.Database.Port)
	}

	if cfg.Redis.Port != "6379" {
		t.Errorf("unexpected redis port: %s", cfg.Redis.Port)
	}
}

func TestConfigInvalidPorts(t *testing.T) {
	tests := []struct {
		name  string
		app   string
		db    string
		redis string
	}{
		{"invalid app port", "abc", "5432", "6379"},
		{"invalid database port", "8080", "abc", "6379"},
		{"invalid redis port", "8080", "5432", "abc"},
		{"zero app port", "0", "5432", "6379"},
		{"app port too high", "65536", "5432", "6379"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			cfg.App.Port = tt.app
			cfg.Database.Port = tt.db
			cfg.Redis.Port = tt.redis

			if err := cfg.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestConfigNegativeRedisDB(t *testing.T) {
	cfg := validConfig()
	cfg.Redis.DB = -1

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected negative Redis DB validation error")
	}
}

func TestConfigEmptyAppName(t *testing.T) {
	cfg := validConfig()
	cfg.App.Name = ""

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected empty app name validation error")
	}
}

func TestConfigInvalidAppEnvironment(t *testing.T) {
	cfg := validConfig()
	cfg.App.Env = "staging"

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "APP_ENV") {
		t.Fatalf("expected APP_ENV validation error, got: %v", err)
	}
}

func TestConfigInvalidCORSOrigins(t *testing.T) {
	tests := []struct {
		name   string
		origin string
	}{
		{"missing scheme", "localhost:5173"},
		{"unsupported scheme", "ftp://localhost:5173"},
		{"path is not an origin", "https://example.com/api"},
		{"wildcard origin", "*"},
		{"query is not allowed", "https://example.com?x=1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			cfg.App.CORSAllowedOrigins = []string{tt.origin}

			err := cfg.Validate()
			if err == nil || !strings.Contains(err.Error(), "CORS") {
				t.Fatalf("expected CORS validation error, got: %v", err)
			}
		})
	}
}

func TestConfigValidCORSOrigins(t *testing.T) {
	cfg := validConfig()
	cfg.App.CORSAllowedOrigins = []string{
		"http://localhost:5173",
		"https://app.example.com",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected CORS validation error: %v", err)
	}
}

func TestConfigInvalidRateLimit(t *testing.T) {
	tests := []struct {
		name   string
		limit  int
		window int
	}{
		{"zero request limit", 0, 60},
		{"negative request limit", -1, 60},
		{"zero window", 100, 0},
		{"negative window", 100, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			cfg.App.RateLimitRequests = tt.limit
			cfg.App.RateLimitWindowSeconds = tt.window

			if err := cfg.Validate(); err == nil {
				t.Fatal("expected rate limit validation error")
			}
		})
	}
}

func TestConfigInvalidRedisPool(t *testing.T) {
	tests := []struct {
		name     string
		poolSize int
		minIdle  int
	}{
		{"zero pool size", 0, 0},
		{"negative minimum idle", 10, -1},
		{"minimum idle exceeds pool", 5, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			cfg.Redis.PoolSize = tt.poolSize
			cfg.Redis.MinIdleConns = tt.minIdle

			if err := cfg.Validate(); err == nil {
				t.Fatal("expected Redis pool validation error")
			}
		})
	}
}

func TestValidateEnvValuesRejectsInvalidIntegers(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"rate limit", "RATE_LIMIT_REQUESTS", "abc"},
		{"database pool", "DATABASE_MAX_OPEN_CONNS", "invalid"},
		{"redis pool", "REDIS_POOL_SIZE", ""},
		{"request timeout", "REQUEST_TIMEOUT_SECONDS", "1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearConfigEnv(t)
			t.Setenv(tt.key, tt.value)

			err := validateEnvValues()
			if err == nil || !strings.Contains(err.Error(), tt.key) {
				t.Fatalf("expected error mentioning %s, got: %v", tt.key, err)
			}
		})
	}
}

func TestValidateEnvValuesRejectsInvalidBoolean(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("TRUST_PROXY", "sometimes")

	err := validateEnvValues()
	if err == nil || !strings.Contains(err.Error(), "TRUST_PROXY") {
		t.Fatalf("expected TRUST_PROXY validation error, got: %v", err)
	}
}

func TestValidateEnvValuesAcceptsValidValues(t *testing.T) {
	clearConfigEnv(t)

	t.Setenv("RATE_LIMIT_REQUESTS", "100")
	t.Setenv("REDIS_POOL_SIZE", "10")
	t.Setenv("TRUST_PROXY", "false")

	if err := validateEnvValues(); err != nil {
		t.Fatalf("unexpected environment validation error: %v", err)
	}
}
