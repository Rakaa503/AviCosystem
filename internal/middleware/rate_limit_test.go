package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func setupRateLimitTest(
	t *testing.T,
	limit int,
) (*miniredis.Miniredis, *redis.Client, *fiber.App) {
	t.Helper()

	miniRedis := miniredis.RunT(t)

	client := redis.NewClient(&redis.Options{
		Addr:         miniRedis.Addr(),
		MaxRetries:   -1,
		DialTimeout:  100 * time.Millisecond,
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
	})

	t.Cleanup(func() {
		_ = client.Close()
	})

	app := fiber.New()
	app.Use(RateLimit(client, limit, time.Minute))

	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	return miniRedis, client, app
}

func sendRateLimitRequest(
	t *testing.T,
	app *fiber.App,
) *http.Response {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/test", nil)

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}

	t.Cleanup(func() {
		_ = response.Body.Close()
	})

	return response
}

func TestRateLimitAllowsRequestsWithinLimit(t *testing.T) {
	_, _, app := setupRateLimitTest(t, 2)

	for i := 1; i <= 2; i++ {
		response := sendRateLimitRequest(t, app)

		if response.StatusCode != fiber.StatusOK {
			t.Fatalf(
				"request %d: expected status 200, got %d",
				i,
				response.StatusCode,
			)
		}

		expectedRemaining := strconv.Itoa(2 - i)

		if got := response.Header.Get("X-RateLimit-Remaining"); got != expectedRemaining {
			t.Errorf(
				"request %d: expected remaining %s, got %s",
				i,
				expectedRemaining,
				got,
			)
		}
	}
}

func TestRateLimitBlocksRequestsOverLimit(t *testing.T) {
	_, _, app := setupRateLimitTest(t, 2)

	sendRateLimitRequest(t, app)
	sendRateLimitRequest(t, app)

	response := sendRateLimitRequest(t, app)

	if response.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf(
			"expected status 429, got %d",
			response.StatusCode,
		)
	}

	if got := response.Header.Get("Retry-After"); got == "" {
		t.Error("expected Retry-After header")
	}

	if got := response.Header.Get("X-RateLimit-Remaining"); got != "0" {
		t.Errorf("expected remaining 0, got %s", got)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if len(body) == 0 {
		t.Error("expected an error response body")
	}
}

func TestRateLimitResetsAfterWindow(t *testing.T) {
	miniRedis, _, app := setupRateLimitTest(t, 1)

	first := sendRateLimitRequest(t, app)
	if first.StatusCode != fiber.StatusOK {
		t.Fatalf("expected first request 200, got %d", first.StatusCode)
	}

	blocked := sendRateLimitRequest(t, app)
	if blocked.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("expected second request 429, got %d", blocked.StatusCode)
	}

	miniRedis.FastForward(time.Minute)

	afterReset := sendRateLimitRequest(t, app)
	if afterReset.StatusCode != fiber.StatusOK {
		t.Fatalf(
			"expected request after reset 200, got %d",
			afterReset.StatusCode,
		)
	}

	if got := afterReset.Header.Get("X-RateLimit-Remaining"); got != "0" {
		t.Errorf("expected remaining 0 after reset, got %s", got)
	}
}

func TestRateLimitFailsClosedWhenRedisUnavailable(t *testing.T) {
	miniRedis, _, app := setupRateLimitTest(t, 2)

	miniRedis.Close()

	response := sendRateLimitRequest(t, app)

	if response.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf(
			"expected status 503 when Redis is unavailable, got %d",
			response.StatusCode,
		)
	}
}
