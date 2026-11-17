package database

import (
	"fmt"
	"time"

	"avicosystem/backend/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgreSQL struct {
	DB *gorm.DB
}

func NewPostgreSQL(cfg config.DatabaseConfig) (*PostgreSQL, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect PostgreSQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL connection: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(
		time.Duration(cfg.ConnMaxLifetimeSeconds) * time.Second,
	)
	sqlDB.SetConnMaxIdleTime(
		time.Duration(cfg.ConnMaxIdleTimeSeconds) * time.Second,
	)

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("PostgreSQL ping failed: %w", err)
	}

	return &PostgreSQL{
		DB: db,
	}, nil
}

func (p *PostgreSQL) Close() error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
