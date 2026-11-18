package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	if err := validateEnvValues(); err != nil {
		return nil, err
	}

	cfg := &Config{
		App:      loadAppConfig(),
		Database: loadDatabaseConfig(),
		Redis:    loadRedisConfig(),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateEnvValues() error {
	integerKeys := []string{
		"APP_PORT",
		"RATE_LIMIT_REQUESTS",
		"RATE_LIMIT_WINDOW_SECONDS",
		"REQUEST_TIMEOUT_SECONDS",
		"JWT_ACCESS_TOKEN_TTL_MINUTES",
		"DATABASE_PORT",
		"DATABASE_MAX_OPEN_CONNS",
		"DATABASE_MAX_IDLE_CONNS",
		"DATABASE_CONN_MAX_LIFETIME_SECONDS",
		"DATABASE_CONN_MAX_IDLE_TIME_SECONDS",
		"REDIS_PORT",
		"REDIS_DB",
		"REDIS_POOL_SIZE",
		"REDIS_MIN_IDLE_CONNS",
		"REDIS_MAX_RETRIES",
		"REDIS_DIAL_TIMEOUT_SECONDS",
		"REDIS_READ_TIMEOUT_SECONDS",
		"REDIS_WRITE_TIMEOUT_SECONDS",
	}

	for _, key := range integerKeys {
		value, exists := os.LookupEnv(key)
		if !exists {
			continue
		}

		if _, err := strconv.Atoi(strings.TrimSpace(value)); err != nil {
			return fmt.Errorf("%s must be a valid integer", key)
		}
	}

	if value, exists := os.LookupEnv("TRUST_PROXY"); exists {
		if _, err := strconv.ParseBool(strings.TrimSpace(value)); err != nil {
			return fmt.Errorf("TRUST_PROXY must be a valid boolean")
		}
	}

	return nil
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.App.Name) == "" {
		return fmt.Errorf("APP_NAME cannot be empty")
	}

	switch c.App.Env {
	case "development", "production", "test":
	default:
		return fmt.Errorf("APP_ENV must be development, production, or test")
	}

	if c.App.Port == "" {
		return fmt.Errorf("APP_PORT cannot be empty")
	}

	if strings.TrimSpace(c.Database.Host) == "" {
		return fmt.Errorf("DATABASE_HOST cannot be empty")
	}

	if strings.TrimSpace(c.Database.User) == "" {
		return fmt.Errorf("DATABASE_USER cannot be empty")
	}

	if strings.TrimSpace(c.Database.Name) == "" {
		return fmt.Errorf("DATABASE_NAME cannot be empty")
	}

	if strings.TrimSpace(c.Redis.Host) == "" {
		return fmt.Errorf("REDIS_HOST cannot be empty")
	}

	if c.Redis.DB < 0 {
		return fmt.Errorf("REDIS_DB cannot be negative")
	}

	ports := map[string]string{
		"APP_PORT":      c.App.Port,
		"DATABASE_PORT": c.Database.Port,
		"REDIS_PORT":    c.Redis.Port,
	}

	for key, value := range ports {
		port, err := strconv.Atoi(value)
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("%s must be a valid port between 1 and 65535", key)
		}
	}

	if c.App.RateLimitRequests <= 0 {
		return fmt.Errorf("RATE_LIMIT_REQUESTS must be greater than zero")
	}

	if c.App.RateLimitWindowSeconds <= 0 {
		return fmt.Errorf("RATE_LIMIT_WINDOW_SECONDS must be greater than zero")
	}

	if c.App.RequestTimeoutSeconds <= 0 {
		return fmt.Errorf("REQUEST_TIMEOUT_SECONDS must be greater than zero")
	}

	if len([]byte(c.App.JWTSecret)) < 32 {
		return fmt.Errorf("JWT_SECRET must contain at least 32 bytes")
	}

	if c.App.JWTAccessTokenTTLMinutes <= 0 {
		return fmt.Errorf("JWT_ACCESS_TOKEN_TTL_MINUTES must be greater than zero")
	}

	if c.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("DATABASE_MAX_OPEN_CONNS must be greater than zero")
	}

	if c.Database.MaxIdleConns < 0 {
		return fmt.Errorf("DATABASE_MAX_IDLE_CONNS cannot be negative")
	}

	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf("DATABASE_MAX_IDLE_CONNS cannot exceed DATABASE_MAX_OPEN_CONNS")
	}

	if c.Database.ConnMaxLifetimeSeconds <= 0 {
		return fmt.Errorf("DATABASE_CONN_MAX_LIFETIME_SECONDS must be greater than zero")
	}

	if c.Database.ConnMaxIdleTimeSeconds <= 0 {
		return fmt.Errorf("DATABASE_CONN_MAX_IDLE_TIME_SECONDS must be greater than zero")
	}

	if c.Redis.PoolSize <= 0 {
		return fmt.Errorf("REDIS_POOL_SIZE must be greater than zero")
	}

	if c.Redis.MinIdleConns < 0 {
		return fmt.Errorf("REDIS_MIN_IDLE_CONNS cannot be negative")
	}

	if c.Redis.MinIdleConns > c.Redis.PoolSize {
		return fmt.Errorf("REDIS_MIN_IDLE_CONNS cannot exceed REDIS_POOL_SIZE")
	}

	if c.Redis.MaxRetries < 0 {
		return fmt.Errorf("REDIS_MAX_RETRIES cannot be negative")
	}

	if c.Redis.DialTimeout <= 0 {
		return fmt.Errorf("REDIS_DIAL_TIMEOUT_SECONDS must be greater than zero")
	}

	if c.Redis.ReadTimeout <= 0 {
		return fmt.Errorf("REDIS_READ_TIMEOUT_SECONDS must be greater than zero")
	}

	if c.Redis.WriteTimeout <= 0 {
		return fmt.Errorf("REDIS_WRITE_TIMEOUT_SECONDS must be greater than zero")
	}

	if err := validateCORSOrigins(c.App.CORSAllowedOrigins); err != nil {
		return err
	}

	if c.App.TrustProxy && len(c.App.TrustedProxies) == 0 {
		return fmt.Errorf("TRUSTED_PROXIES must contain at least one IP or CIDR when TRUST_PROXY is enabled")
	}

	for _, proxy := range c.App.TrustedProxies {
		if net.ParseIP(proxy) != nil {
			continue
		}

		if _, _, err := net.ParseCIDR(proxy); err != nil {
			return fmt.Errorf("invalid trusted proxy address or CIDR: %s", proxy)
		}
	}

	return nil
}

func validateCORSOrigins(origins []string) error {
	for _, origin := range origins {
		parsed, err := url.Parse(origin)
		if err != nil {
			return fmt.Errorf("invalid CORS origin: %s", origin)
		}

		if (parsed.Scheme != "http" && parsed.Scheme != "https") ||
			parsed.Host == "" ||
			parsed.User != nil ||
			parsed.Path != "" ||
			parsed.RawQuery != "" ||
			parsed.Fragment != "" {
			return fmt.Errorf("invalid CORS origin: %s", origin)
		}
	}

	return nil
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}

	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}

	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}

	return parsed
}
