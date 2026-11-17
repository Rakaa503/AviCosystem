package server

import (
	"avicosystem/backend/internal/cache"
	"avicosystem/backend/internal/config"
	"avicosystem/backend/internal/database"

	"github.com/gofiber/fiber/v3"
)

func New(
	cfg *config.Config,
	postgres *database.PostgreSQL,
	redisClient *cache.Redis,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: cfg.App.Name,
	})

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

	return app
}
