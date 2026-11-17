package cache

import (
	"testing"

	"avicosystem/backend/internal/config"
	"github.com/joho/godotenv"
)

func TestRedisConnection(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("Failed to load .env: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	redisClient, err := NewRedis(cfg.Redis)
	if err != nil {
		t.Fatalf("Redis connection failed: %v", err)
	}

	defer func() {
		if err := redisClient.Close(); err != nil {
			t.Errorf("Failed to close Redis: %v", err)
		}
	}()

	result, err := redisClient.Client.Ping(t.Context()).Result()
	if err != nil {
		t.Fatalf("Redis ping failed: %v", err)
	}

	if result != "PONG" {
		t.Fatalf("Unexpected Redis response: %s", result)
	}

	t.Logf("Redis connection successful: %s", result)
	t.Logf("Redis pool size: %d", cfg.Redis.PoolSize)
	t.Logf("Redis minimum idle connections: %d", cfg.Redis.MinIdleConns)
}
