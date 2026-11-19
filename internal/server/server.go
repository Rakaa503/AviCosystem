package server

import (
	"time"

	"avicosystem/backend/internal/cache"
	"avicosystem/backend/internal/config"
	"avicosystem/backend/internal/database"
	appmiddleware "avicosystem/backend/internal/middleware"
	authmodule "avicosystem/backend/internal/modules/auth"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	fiberTimeout "github.com/gofiber/fiber/v3/middleware/timeout"
)

func New(
	cfg *config.Config,
	postgres *database.PostgreSQL,
	redisClient *cache.Redis,
	authHandler *authmodule.Handler,
) *fiber.App {
	fiberConfig := fiber.Config{
		AppName:    cfg.App.Name,
		TrustProxy: cfg.App.TrustProxy,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies: cfg.App.TrustedProxies,
		},
		EnableIPValidation: true,
	}

	if cfg.App.TrustProxy {
		fiberConfig.ProxyHeader = "X-Forwarded-For"
	}

	app := fiber.New(fiberConfig)

	app.Use(recover.New())
	app.Use(appmiddleware.RequestID())
	app.Use(appmiddleware.RequestLogger())
	app.Use(appmiddleware.SecurityHeaders())
	app.Use(appmiddleware.CORS(cfg.App.CORSAllowedOrigins))

	app.Use(appmiddleware.RateLimit(
		redisClient.Client,
		cfg.App.RateLimitRequests,
		time.Duration(cfg.App.RateLimitWindowSeconds)*time.Second,
	))

	app.Use(fiberTimeout.New(
		func(c fiber.Ctx) error {
			return c.Next()
		},
		fiberTimeout.Config{
			Timeout: time.Duration(cfg.App.RequestTimeoutSeconds) * time.Second,
			OnTimeout: func(c fiber.Ctx) error {
				return c.Status(fiber.StatusServiceUnavailable).JSON(map[string]any{
					"status":  "error",
					"message": "Request processing timeout",
				})
			},
		},
	))

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(map[string]any{
			"status":  "ok",
			"service": cfg.App.Name,
		})
	})

	app.Get("/ready", func(c fiber.Ctx) error {
		sqlDB, err := postgres.DB.DB()
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(map[string]any{
				"status": "not_ready",
				"error":  "PostgreSQL unavailable",
			})
		}

		if err := sqlDB.PingContext(c.Context()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(map[string]any{
				"status": "not_ready",
				"error":  "PostgreSQL unavailable",
			})
		}

		if err := redisClient.Client.Ping(c.Context()).Err(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(map[string]any{
				"status": "not_ready",
				"error":  "Redis unavailable",
			})
		}

		return c.JSON(map[string]any{
			"status": "ready",
			"dependencies": map[string]string{
				"postgresql": "connected",
				"redis":      "connected",
			},
		})
	})

	authRoutes := app.Group("/api/v1/auth")
	authRoutes.Post("/register", authHandler.Register)
	authRoutes.Post("/login", authHandler.Login)

	return app
}
