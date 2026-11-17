package config

import (
	"os"
	"testing"
)

func clearConfigEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"APP_NAME", "APP_ENV", "APP_PORT",
		"DATABASE_HOST", "DATABASE_PORT",
		"DATABASE_USER", "DATABASE_PASSWORD",
		"DATABASE_NAME", "DATABASE_SSLMODE",
		"REDIS_HOST", "REDIS_PORT",
		"REDIS_PASSWORD", "REDIS_DB",
	}

	for _, key := range keys {
		t.Setenv(key, "")
		os.Unsetenv(key)
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
			cfg := &Config{
				App: AppConfig{
					Name: "AviCosystem",
					Port: tt.app,
				},
				Database: DatabaseConfig{
					Host: "localhost",
					Port: tt.db,
					User: "postgres",
					Name: "avicosystem",
				},
				Redis: RedisConfig{
					Host: "localhost",
					Port: tt.redis,
				},
			}

			if err := cfg.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestConfigNegativeRedisDB(t *testing.T) {
	cfg := &Config{
		App: AppConfig{
			Name: "AviCosystem",
			Port: "8080",
		},
		Database: DatabaseConfig{
			Host: "localhost",
			Port: "5432",
			User: "postgres",
			Name: "avicosystem",
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: "6379",
			DB:   -1,
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected negative Redis DB validation error")
	}
}

func TestConfigEmptyAppName(t *testing.T) {
	cfg := &Config{
		App: AppConfig{
			Name: "",
			Port: "8080",
		},
		Database: DatabaseConfig{
			Host: "localhost",
			Port: "5432",
			User: "postgres",
			Name: "avicosystem",
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: "6379",
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected empty app name validation error")
	}
}
