package database

import (
	"testing"

	"avicosystem/backend/internal/config"
	"github.com/joho/godotenv"
)

func TestPostgreSQLConnection(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		t.Logf("No .env file found; using environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := NewPostgreSQL(cfg.Database)
	if err != nil {
		t.Fatalf("PostgreSQL connection failed: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	sqlDB, err := db.DB.DB()
	if err != nil {
		t.Fatalf("Failed to retrieve SQL connection: %v", err)
	}

	stats := sqlDB.Stats()

	if stats.MaxOpenConnections != cfg.Database.MaxOpenConns {
		t.Fatalf(
			"Unexpected max open connections: got %d, want %d",
			stats.MaxOpenConnections,
			cfg.Database.MaxOpenConns,
		)
	}

	var databaseName string

	if err := db.DB.Raw("SELECT current_database()").Scan(&databaseName).Error; err != nil {
		t.Fatalf("Database query failed: %v", err)
	}

	if databaseName != cfg.Database.Name {
		t.Fatalf("Unexpected database: %s", databaseName)
	}

	t.Logf("Connected successfully to database: %s", databaseName)
	t.Logf("Max open connections: %d", stats.MaxOpenConnections)
	t.Logf("Max idle connections: %d", cfg.Database.MaxIdleConns)
}
