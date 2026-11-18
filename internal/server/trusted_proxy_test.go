package server

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func startTrustedProxyTestServer(
	t *testing.T,
	trustedProxies []string,
) string {
	t.Helper()

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

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create TCP listener: %v", err)
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

	return fmt.Sprintf("http://%s/client-ip", address)
}

func requestClientIP(
	t *testing.T,
	url string,
	forwardedFor string,
) string {
	t.Helper()

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if forwardedFor != "" {
		request.Header.Set("X-Forwarded-For", forwardedFor)
	}

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	response, err := client.Do(request)
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
	url := startTrustedProxyTestServer(
		t,
		[]string{"127.0.0.1"},
	)

	actual := requestClientIP(t, url, "203.0.113.10")

	if actual != "203.0.113.10" {
		t.Fatalf("expected client IP 203.0.113.10, got %s", actual)
	}
}

func TestUntrustedProxyCannotOverrideClientIP(t *testing.T) {
	url := startTrustedProxyTestServer(
		t,
		[]string{"10.0.0.0/8"},
	)

	actual := requestClientIP(t, url, "203.0.113.99")

	if actual != "127.0.0.1" {
		t.Fatalf("expected remote IP 127.0.0.1, got %s", actual)
	}
}

func TestTrustedProxySkipsTrustedIPsInForwardedChain(t *testing.T) {
	url := startTrustedProxyTestServer(
		t,
		[]string{
			"127.0.0.1",
			"10.0.0.0/8",
		},
	)

	actual := requestClientIP(
		t,
		url,
		"203.0.113.15, 10.0.0.1",
	)

	if actual != "203.0.113.15" {
		t.Fatalf("expected client IP 203.0.113.15, got %s", actual)
	}
}
