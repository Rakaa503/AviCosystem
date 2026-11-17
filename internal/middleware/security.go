package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

func SecurityHeaders() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Set("X-XSS-Protection", "0")

		return c.Next()
	}
}

func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		idBytes := make([]byte, 16)

		if _, err := rand.Read(idBytes); err != nil {
			return fiber.NewError(
				fiber.StatusInternalServerError,
				"Failed to generate request ID",
			)
		}

		requestID := hex.EncodeToString(idBytes)

		c.Locals("request_id", requestID)
		c.Set("X-Request-ID", requestID)

		return c.Next()
	}
}

func RequestLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		requestID, _ := c.Locals("request_id").(string)

		log.Printf(
			"request_id=%s method=%s path=%s status=%d latency=%s",
			requestID,
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			time.Since(start),
		)

		return err
	}
}

func CORS(allowedOrigins []string) fiber.Handler {
	return func(c fiber.Ctx) error {
		origin := c.Get("Origin")

		if origin == "" {
			return c.Next()
		}

		allowed := false

		for _, allowedOrigin := range allowedOrigins {
			if strings.TrimSpace(allowedOrigin) == origin {
				allowed = true
				break
			}
		}

		if !allowed {
			return fiber.NewError(
				fiber.StatusForbidden,
				"Origin not allowed",
			)
		}

		c.Set("Access-Control-Allow-Origin", origin)
		c.Set("Vary", "Origin")
		c.Set(
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, PATCH, DELETE, OPTIONS",
		)
		c.Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Authorization, X-Request-ID",
		)
		c.Set("Access-Control-Expose-Headers", "X-Request-ID")
		c.Set("Access-Control-Max-Age", "600")

		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}
