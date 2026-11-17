package database

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestPostgreSQLConnection(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("Failed to load .env: %v", err)
	}

	required := []string{
		"DATABASE_HOST",
		"DATABASE_PORT",
		"DATABASE_USER",
		"DATABASE_PASSWORD",
		"DATABASE_NAME",
		"DATABASE_SSLMODE",
	}

	for _, key := range required {
		if os.Getenv(key) == "" {
			t.Fatalf("Missing environment variable: %s", key)
		}
	}

	db, err := NewPostgreSQL(
		os.Getenv("DATABASE_HOST"),
		os.Getenv("DATABASE_PORT"),
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_NAME"),
		os.Getenv("DATABASE_SSLMODE"),
	)

	if err != nil {
		t.Fatalf("PostgreSQL connection failed: %v", err)
	}

	defer db.Close()

	var databaseName string

	if err := db.DB.Raw("SELECT current_database()").Scan(&databaseName).Error; err != nil {
		t.Fatalf("Database query failed: %v", err)
	}

	if databaseName != "avicosystem" {
		t.Fatalf("Unexpected database: %s", databaseName)
	}

	t.Logf("Connected successfully to database: %s", databaseName)
}
