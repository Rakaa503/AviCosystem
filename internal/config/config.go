package config

import (
	"fmt"
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
