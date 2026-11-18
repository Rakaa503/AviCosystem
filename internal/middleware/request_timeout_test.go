package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	fiberTimeout "github.com/gofiber/fiber/v3/middleware/timeout"
)

func newTimeoutTestApp(timeoutDuration time.Duration) *fiber.App {
	app := fiber.New()

	app.Use(fiberTimeout.New(
		func(c fiber.Ctx) error {
			return c.Next()
		},
		fiberTimeout.Config{
			Timeout: timeoutDuration,
			OnTimeout: func(c fiber.Ctx) error {
				return c.Status(fiber.StatusServiceUnavailable).JSON(map[string]string{
					"status":  "error",
					"message": "Request processing timeout",
				})
			},
		},
	))

	return app
}

func TestRequestTimeoutAllowsNormalRequest(t *testing.T) {
	app := newTimeoutTestApp(200 * time.Millisecond)

	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	response, err := app.Test(
		httptest.NewRequest(http.MethodGet, "/health", nil),
	)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}
}

func TestRequestTimeoutReturnsServiceUnavailable(t *testing.T) {
	app := newTimeoutTestApp(50 * time.Millisecond)

	cancelled := make(chan struct{})

	app.Get("/slow", func(c fiber.Ctx) error {
		<-c.Context().Done()
		close(cancelled)
		return nil
	})

	response, err := app.Test(
		httptest.NewRequest(http.MethodGet, "/slow", nil),
	)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status 503, got %d",
			response.StatusCode,
		)
	}

	select {
	case <-cancelled:
		// Handler received context cancellation.
	case <-time.After(time.Second):
		t.Fatal("handler did not receive context cancellation")
	}
}
