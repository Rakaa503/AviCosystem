package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func newTrustedProxyTestApp(trustedProxies []string) *fiber.App {
	app := fiber.New(fiber.Config{
		TrustProxy: true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Proxies: trustedProxies,
		},
		ProxyHeader:        "X-Forwarded-For",
		EnableIPValidation: true,
	})

	app.Get("/client-ip", func(c fiber.Ctx) error {
		return c.SendString(c.IP())
	})

	return app
}

func performIPRequest(
	t *testing.T,
	app *fiber.App,
	remoteAddr string,
	forwardedFor string,
) string {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/client-ip", nil)
	request.RemoteAddr = remoteAddr

	if forwardedFor != "" {
		request.Header.Set("X-Forwarded-For", forwardedFor)
	}

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	return string(body)
}

func TestTrustedProxyUsesForwardedClientIP(t *testing.T) {
	app := newTrustedProxyTestApp([]string{"10.0.0.0/8"})

	actual := performIPRequest(
		t,
		app,
		"10.0.0.2:8080",
		"203.0.113.10, 10.0.0.1",
	)

	if actual != "203.0.113.10" {
		t.Fatalf("expected client IP 203.0.113.10, got %s", actual)
	}
}

func TestUntrustedProxyCannotOverrideClientIP(t *testing.T) {
	app := newTrustedProxyTestApp([]string{"10.0.0.0/8"})

	actual := performIPRequest(
		t,
		app,
		"198.51.100.20:8080",
		"203.0.113.99",
	)

	if actual != "198.51.100.20" {
		t.Fatalf("expected remote IP 198.51.100.20, got %s", actual)
	}
}

func TestTrustedProxySkipsTrustedIPsInForwardedChain(t *testing.T) {
	app := newTrustedProxyTestApp([]string{"10.0.0.0/8"})

	actual := performIPRequest(
		t,
		app,
		"10.0.0.2:8080",
		"203.0.113.15, 10.0.0.1",
	)

	if actual != "203.0.113.15" {
		t.Fatalf("expected client IP 203.0.113.15, got %s", actual)
	}
}
