package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"avicosystem/backend/internal/cache"
	"avicosystem/backend/internal/config"
	"avicosystem/backend/internal/database"
	authmodule "avicosystem/backend/internal/modules/auth"
	"avicosystem/backend/internal/modules/users"
	"avicosystem/backend/internal/server"
)

func main() {
	if err := run(); err != nil {
		log.Printf("Application stopped: %v", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("configuration initialization failed: %w", err)
	}

	postgres, err := database.NewPostgreSQL(cfg.Database)
	if err != nil {
		return fmt.Errorf("PostgreSQL initialization failed: %w", err)
	}
	defer postgres.Close()

	redisClient, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		return fmt.Errorf("Redis initialization failed: %w", err)
	}
	defer redisClient.Close()

	jwtService, err := authmodule.NewJWTService(
		cfg.App.JWTSecret,
		cfg.App.Name,
		time.Duration(cfg.App.JWTAccessTokenTTLMinutes)*time.Minute,
	)
	if err != nil {
		return fmt.Errorf("JWT initialization failed: %w", err)
	}

	userRepository := users.NewRepository(postgres.DB)
	passwordService := users.NewPasswordService()

	authService, err := authmodule.NewAuthenticationService(
		userRepository,
		passwordService,
		jwtService,
	)
	if err != nil {
		return fmt.Errorf("authentication initialization failed: %w", err)
	}

	authHandler := authmodule.NewHandler(authService)

	app := server.New(cfg, postgres, redisClient, authHandler)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverError := make(chan error, 1)

	go func() {
		serverError <- app.Listen(":" + cfg.App.Port)
	}()

	log.Printf(
		"%s running on port %s (%s)",
		cfg.App.Name,
		cfg.App.Port,
		cfg.App.Env,
	)

	select {
	case err := <-serverError:
		if err != nil {
			return fmt.Errorf("HTTP server failed: %w", err)
		}

		return nil

	case <-ctx.Done():
		log.Println("Shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			return fmt.Errorf("HTTP server shutdown failed: %w", err)
		}

		log.Println("HTTP server stopped gracefully")
		return nil
	}
}
