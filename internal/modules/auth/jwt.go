package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidJWTConfig = errors.New("invalid JWT configuration")
	ErrInvalidToken     = errors.New("invalid or expired token")
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

func NewJWTService(
	secret string,
	issuer string,
	ttl time.Duration,
) (*JWTService, error) {
	if len([]byte(secret)) < 32 {
		return nil, fmt.Errorf("%w: secret must contain at least 32 bytes", ErrInvalidJWTConfig)
	}

	if strings.TrimSpace(issuer) == "" {
		return nil, fmt.Errorf("%w: issuer cannot be empty", ErrInvalidJWTConfig)
	}

	if ttl <= 0 {
		return nil, fmt.Errorf("%w: TTL must be positive", ErrInvalidJWTConfig)
	}

	return &JWTService{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    ttl,
	}, nil
}

func (s *JWTService) GenerateAccessToken(
	userID uuid.UUID,
	role string,
) (string, error) {
	if userID == uuid.Nil || strings.TrimSpace(role) == "" {
		return "", ErrInvalidToken
	}

	now := time.Now()

	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(s.secret)
}

func (s *JWTService) ValidateAccessToken(
	tokenString string,
) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidToken
			}

			return s.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(s.issuer),
		jwt.WithExpirationRequired(),
	)

	if err != nil || token == nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	if _, err := uuid.Parse(claims.Subject); err != nil {
		return nil, ErrInvalidToken
	}

	if strings.TrimSpace(claims.Role) == "" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
