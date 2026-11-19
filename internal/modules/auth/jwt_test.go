package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	testJWTSecret = "test-secret-with-at-least-32-bytes-long"
	testJWTIssuer = "AviCosystem"
)

func newTestJWTService(t *testing.T) *JWTService {
	t.Helper()

	service, err := NewJWTService(
		testJWTSecret,
		testJWTIssuer,
		15*time.Minute,
	)
	if err != nil {
		t.Fatalf("failed to create JWT service: %v", err)
	}

	return service
}

func TestJWTServiceGenerateAndValidate(t *testing.T) {
	service := newTestJWTService(t)
	userID := uuid.New()

	token, err := service.GenerateAccessToken(userID, "user")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := service.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.Subject != userID.String() {
		t.Fatalf("unexpected subject: %s", claims.Subject)
	}

	if claims.Role != "user" {
		t.Fatalf("unexpected role: %s", claims.Role)
	}

	if claims.Issuer != testJWTIssuer {
		t.Fatalf("unexpected issuer: %s", claims.Issuer)
	}
}

func TestJWTServiceRejectsTamperedToken(t *testing.T) {
	service := newTestJWTService(t)

	token, err := service.GenerateAccessToken(uuid.New(), "user")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	tamperedToken := token[:len(token)-1] + "x"

	_, err = service.ValidateAccessToken(tamperedToken)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got: %v", err)
	}
}

func TestJWTServiceRejectsExpiredToken(t *testing.T) {
	service := newTestJWTService(t)

	claims := Claims{
		Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    testJWTIssuer,
			Subject:   uuid.New().String(),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(service.secret)
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	_, err = service.ValidateAccessToken(signedToken)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got: %v", err)
	}
}

func TestJWTServiceRejectsWrongIssuer(t *testing.T) {
	service := newTestJWTService(t)

	claims := Claims{
		Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "WrongIssuer",
			Subject:   uuid.New().String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(service.secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = service.ValidateAccessToken(signedToken)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got: %v", err)
	}
}

func TestJWTServiceRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		secret string
		issuer string
		ttl    time.Duration
	}{
		{"short secret", "short", testJWTIssuer, time.Minute},
		{"empty issuer", testJWTSecret, "", time.Minute},
		{"zero TTL", testJWTSecret, testJWTIssuer, 0},
		{"negative TTL", testJWTSecret, testJWTIssuer, -time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewJWTService(tt.secret, tt.issuer, tt.ttl)
			if !errors.Is(err, ErrInvalidJWTConfig) {
				t.Fatalf("expected ErrInvalidJWTConfig, got: %v", err)
			}
		})
	}
}

func TestJWTServiceRejectsInvalidUserData(t *testing.T) {
	service := newTestJWTService(t)

	tests := []struct {
		name   string
		userID uuid.UUID
		role   string
	}{
		{"nil user ID", uuid.Nil, "user"},
		{"empty role", uuid.New(), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GenerateAccessToken(tt.userID, tt.role)
			if !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("expected ErrInvalidToken, got: %v", err)
			}
		})
	}
}
