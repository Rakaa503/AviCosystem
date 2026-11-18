package server

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	appmiddleware "avicosystem/backend/internal/middleware"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

func startRateLimitProxyTestServer(
	t *testing.T,
	trustedProxies []string,
	limit int,
) string {
	t.Helper()

	miniRedis := miniredis.RunT(t)

	redisClient := redis.NewClient(&redis.Options{
		Addr:         miniRedis.Addr(),
		MaxRetries:   -1,
		DialTimeout:  100 * time.Millisecond,
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
	})

	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	app := fiber.New(fiber.Config{
		TrustProxy: true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies: trustedProxies,
		},
		ProxyHeader:        "X-Forwarded-For",
		EnableIPValidation: true,
	})

	app.Use(appmiddleware.RateLimit(
		redisClient,
		limit,
		time.Minute,
	))

	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	address := listener.Addr().String()
	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- app.Listener(listener)
	}()

	t.Cleanup(func() {
		_ = app.Shutdown()
		_ = listener.Close()

		select {
		case err := <-serverErrors:
			if err != nil {
				t.Logf("listener stopped: %v", err)
			}
		case <-time.After(time.Second):
			t.Log("listener shutdown timed out")
		}
	})

	return fmt.Sprintf("http://%s/test", address)
}

func sendProxyRateLimitRequest(
	t *testing.T,
	url string,
	forwardedFor string,
) int {
	t.Helper()

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	request.Header.Set("X-Forwarded-For", forwardedFor)

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer response.Body.Close()

	return response.StatusCode
}

func TestRateLimitUsesTrustedForwardedClientIP(t *testing.T) {
	url := startRateLimitProxyTestServer(
		t,
		[]string{"127.0.0.1"},
		1,
	)

	first := sendProxyRateLimitRequest(
		t,
		url,
		"203.0.113.10",
	)

	second := sendProxyRateLimitRequest(
		t,
		url,
		"203.0.113.10",
	)

	if first != fiber.StatusOK {
		t.Fatalf("expected first request 200, got %d", first)
	}

	if second != fiber.StatusTooManyRequests {
		t.Fatalf("expected second request 429, got %d", second)
	}
}

func TestRateLimitRejectsUntrustedForwardedIPSpoofing(t *testing.T) {
	url := startRateLimitProxyTestServer(
		t,
		[]string{"10.0.0.0/8"},
		1,
	)

	first := sendProxyRateLimitRequest(
		t,
		url,
		"203.0.113.10",
	)

	second := sendProxyRateLimitRequest(
		t,
		url,
		"203.0.113.11",
	)

	if first != fiber.StatusOK {
		t.Fatalf("expected first request 200, got %d", first)
	}

	if second != fiber.StatusTooManyRequests {
		t.Fatalf(
			"expected spoofed request to share original IP bucket and return 429, got %d",
			second,
		)
	}
}

func TestRateLimitSeparatesTrustedForwardedClientIPs(t *testing.T) {
	url := startRateLimitProxyTestServer(
		t,
		[]string{"127.0.0.1"},
		1,
	)

	firstClient := sendProxyRateLimitRequest(
		t,
		url,
		"203.0.113.10",
	)

	secondClient := sendProxyRateLimitRequest(
		t,
		url,
		"203.0.113.11",
	)

	repeatedFirstClient := sendProxyRateLimitRequest(
		t,
		url,
		"203.0.113.10",
	)

	if firstClient != fiber.StatusOK {
		t.Fatalf("expected first client 200, got %d", firstClient)
	}

	if secondClient != fiber.StatusOK {
		t.Fatalf("expected second client 200, got %d", secondClient)
	}

	if repeatedFirstClient != fiber.StatusTooManyRequests {
		t.Fatalf(
			"expected repeated first client 429, got %d",
			repeatedFirstClient,
		)
	}
}
