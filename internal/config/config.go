package config

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

func Load() (*Config, error) {
	_ = godotenv.Load()

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

func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("APP_NAME cannot be empty")
	}

	if c.App.Port == "" {
		return fmt.Errorf("APP_PORT cannot be empty")
	}

	if c.Database.Host == "" {
		return fmt.Errorf("DATABASE_HOST cannot be empty")
	}

	if c.Database.User == "" {
		return fmt.Errorf("DATABASE_USER cannot be empty")
	}

	if c.Database.Name == "" {
		return fmt.Errorf("DATABASE_NAME cannot be empty")
	}

	if c.Redis.Host == "" {
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

	if c.App.RequestTimeoutSeconds <= 0 {
		return fmt.Errorf("REQUEST_TIMEOUT_SECONDS must be greater than zero")
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

	parsed, err := strconv.Atoi(value)
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

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}
