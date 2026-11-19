package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"avicosystem/backend/internal/modules/users"

	"github.com/gofiber/fiber/v3"
)

func setupAuthHandlerTest(t *testing.T) (*fiber.App, *users.Repository, *JWTService) {
	t.Helper()

	service, repository, _ := setupAuthenticationService(t)

	app := fiber.New()
	handler := NewHandler(service)

	routes := app.Group("/api/v1/auth")
	routes.Post("/register", handler.Register)
	routes.Post("/login", handler.Login)

	return app, repository, service.jwt
}

func sendAuthJSON(
	t *testing.T,
	app *fiber.App,
	method string,
	path string,
	body string,
) (*http.Response, string) {
	t.Helper()

	request := httptest.NewRequest(
		method,
		path,
		strings.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}

	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	return response, string(responseBody)
}

func TestAuthHandlerHTTPFlow(t *testing.T) {
	app, repository, jwtService := setupAuthHandlerTest(t)

	registerBody := `{
"name": "Integration User",
"email": "integration@example.com",
"password": "StrongPassword123!"
}`

	t.Run("register success", func(t *testing.T) {
		response, body := sendAuthJSON(
			t,
			app,
			http.MethodPost,
			"/api/v1/auth/register",
			registerBody,
		)

		if response.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", response.StatusCode, body)
		}

		if strings.Contains(body, "password_hash") {
			t.Fatal("response must not expose password_hash")
		}

		if strings.Contains(body, "StrongPassword123!") {
			t.Fatal("response must not expose plaintext password")
		}

		var payload struct {
			Status string `json:"status"`
			Data   struct {
				ID    string `json:"id"`
				Email string `json:"email"`
				Role  string `json:"role"`
			} `json:"data"`
		}

		if err := json.Unmarshal([]byte(body), &payload); err != nil {
			t.Fatalf("invalid registration response: %v", err)
		}

		if payload.Status != "success" {
			t.Fatalf("unexpected response status: %s", payload.Status)
		}

		if payload.Data.Email != "integration@example.com" {
			t.Fatalf("unexpected email: %s", payload.Data.Email)
		}

		if payload.Data.Role != string(users.RoleUser) {
			t.Fatalf("unexpected role: %s", payload.Data.Role)
		}

		if payload.Data.ID == "" {
			t.Fatal("expected user ID in response")
		}
	})

	t.Run("duplicate registration", func(t *testing.T) {
		response, body := sendAuthJSON(
			t,
			app,
			http.MethodPost,
			"/api/v1/auth/register",
			registerBody,
		)

		if response.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d: %s", response.StatusCode, body)
		}
	})

	t.Run("invalid registration", func(t *testing.T) {
		response, body := sendAuthJSON(
			t,
			app,
			http.MethodPost,
			"/api/v1/auth/register",
			`{"name":"","email":"invalid","password":""}`,
		)

		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", response.StatusCode, body)
		}
	})

	t.Run("login invalid password", func(t *testing.T) {
		response, body := sendAuthJSON(
			t,
			app,
			http.MethodPost,
			"/api/v1/auth/login",
			`{"email":"integration@example.com","password":"WrongPassword123!"}`,
		)

		if response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d: %s", response.StatusCode, body)
		}
	})

	t.Run("login success and validate JWT", func(t *testing.T) {
		response, body := sendAuthJSON(
			t,
			app,
			http.MethodPost,
			"/api/v1/auth/login",
			`{"email":"integration@example.com","password":"StrongPassword123!"}`,
		)

		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", response.StatusCode, body)
		}

		if strings.Contains(body, "password_hash") {
			t.Fatal("response must not expose password_hash")
		}
		var payload struct {
			User struct {
				Email string `json:"email"`
			} `json:"user"`
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			ExpiresIn   int64  `json:"expires_in"`
		}

		if err := json.Unmarshal([]byte(body), &payload); err != nil {
			t.Fatalf("invalid login response: %v", err)
		}

		if payload.AccessToken == "" {
			t.Fatal("expected access token")
		}

		if payload.TokenType != "Bearer" {
			t.Fatalf("unexpected token type: %s", payload.TokenType)
		}

		if payload.ExpiresIn <= 0 {
			t.Fatalf("invalid token expiration: %d", payload.ExpiresIn)
		}

		if payload.User.Email != "integration@example.com" {
			t.Fatalf("unexpected user email: %s", payload.User.Email)
		}

		claims, err := jwtService.ValidateAccessToken(payload.AccessToken)
		if err != nil {
			t.Fatalf("generated JWT is invalid: %v", err)
		}

		if claims.Role != string(users.RoleUser) {
			t.Fatalf("unexpected JWT role: %s", claims.Role)
		}
	})

	t.Run("inactive account", func(t *testing.T) {
		passwordService := users.NewPasswordService()

		hash, err := passwordService.HashPassword("InactivePassword123!")
		if err != nil {
			t.Fatalf("failed to hash password: %v", err)
		}

		inactiveUser := &users.User{
			Name:         "Inactive User",
			Email:        "inactive@example.com",
			PasswordHash: hash,
			Role:         users.RoleUser,
			Status:       users.StatusSuspended,
		}

		if err := repository.Create(inactiveUser); err != nil {
			t.Fatalf("failed to create inactive user: %v", err)
		}

		response, body := sendAuthJSON(
			t,
			app,
			http.MethodPost,
			"/api/v1/auth/login",
			`{"email":"inactive@example.com","password":"InactivePassword123!"}`,
		)

		if response.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403, got %d: %s", response.StatusCode, body)
		}
	})
}
