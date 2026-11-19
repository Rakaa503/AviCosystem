package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authmodule "avicosystem/backend/internal/modules/auth"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func setupAuthMiddlewareTest(t *testing.T) (*authmodule.JWTService, *fiber.App) {
	t.Helper()

	jwtService, err := authmodule.NewJWTService(
		"test-secret-with-at-least-32-bytes-long",
		"AviCosystem-Test",
		15*time.Minute,
	)
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}

	app := fiber.New()
	app.Use(Auth(jwtService))

	app.Get("/protected", func(c fiber.Ctx) error {
		userID, ok := c.Locals(LocalUserID).(uuid.UUID)
		if !ok {
			return fiber.NewError(fiber.StatusInternalServerError)
		}

		role, ok := c.Locals(LocalUserRole).(string)
		if !ok {
			return fiber.NewError(fiber.StatusInternalServerError)
		}

		return c.JSON(fiber.Map{
			"user_id": userID.String(),
			"role":    role,
		})
	})

	return jwtService, app
}

func TestAuthMiddleware(t *testing.T) {
	jwtService, app := setupAuthMiddlewareTest(t)

	userID := uuid.New()

	validToken, err := jwtService.GenerateAccessToken(userID, "user")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	tests := []struct {
		name           string
		authorization  string
		expectedStatus int
	}{
		{
			name:           "valid bearer token",
			authorization:  "Bearer " + validToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing authorization header",
			authorization:  "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing bearer scheme",
			authorization:  validToken,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid bearer format",
			authorization:  "Bearer",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "extra authorization value",
			authorization:  "Bearer " + validToken + " extra",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			authorization:  "Bearer invalid.token.value",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "wrong authorization scheme",
			authorization:  "Basic " + validToken,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)

			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.expectedStatus,
					resp.StatusCode,
				)
			}

			if tt.expectedStatus == http.StatusOK {
				body := make([]byte, 1024)
				n, _ := resp.Body.Read(body)
				responseBody := string(body[:n])

				if !strings.Contains(responseBody, userID.String()) {
					t.Fatalf("user ID missing from response: %s", responseBody)
				}

				if !strings.Contains(responseBody, `"role":"user"`) {
					t.Fatalf("user role missing from response: %s", responseBody)
				}
			}
		})
	}
}
