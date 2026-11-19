package auth

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"avicosystem/backend/internal/modules/users"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrInvalidInput       = errors.New("invalid authentication input")
	ErrAccountUnavailable = errors.New("account is unavailable")
)

type UserRepository interface {
	Create(user *users.User) error
	FindByEmail(email string) (*users.User, error)
	UpdateLastLogin(id uuid.UUID, loginAt time.Time) error
}

type PasswordManager interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hashedPassword string) error
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthResult struct {
	User        *users.User `json:"user"`
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int64       `json:"expires_in"`
}

type AuthenticationService struct {
	repository UserRepository
	password   PasswordManager
	jwt        *JWTService
}

func NewAuthenticationService(
	repository UserRepository,
	password PasswordManager,
	jwtService *JWTService,
) (*AuthenticationService, error) {
	if repository == nil || password == nil || jwtService == nil {
		return nil, errors.New("authentication dependencies cannot be nil")
	}

	return &AuthenticationService{
		repository: repository,
		password:   password,
		jwt:        jwtService,
	}, nil
}

func (s *AuthenticationService) Register(
	input RegisterInput,
) (*users.User, error) {
	name := strings.TrimSpace(input.Name)
	email := normalizeAuthEmail(input.Email)

	if name == "" ||
		len(name) > 100 ||
		len(email) > 255 ||
		!validEmail(email) ||
		input.Password == "" {
		return nil, ErrInvalidInput
	}

	if _, err := s.repository.FindByEmail(email); err == nil {
		return nil, ErrEmailAlreadyExists
	} else if !errors.Is(err, users.ErrUserNotFound) {
		return nil, fmt.Errorf("check existing user: %w", err)
	}

	passwordHash, err := s.password.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &users.User{
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         users.RoleUser,
		Status:       users.StatusActive,
	}

	if err := s.repository.Create(user); err != nil {
		return nil, fmt.Errorf("register user: %w", err)
	}

	user.PasswordHash = ""

	return user, nil
}

func (s *AuthenticationService) Login(
	input LoginInput,
) (*AuthResult, error) {
	email := normalizeAuthEmail(input.Email)

	if !validEmail(email) || input.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.repository.FindByEmail(email)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf("find login user: %w", err)
	}

	if user.Status != users.StatusActive {
		return nil, ErrAccountUnavailable
	}

	if err := s.password.VerifyPassword(
		input.Password,
		user.PasswordHash,
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	now := time.Now().UTC()

	token, err := s.jwt.GenerateAccessToken(
		user.ID,
		string(user.Role),
	)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	if err := s.repository.UpdateLastLogin(user.ID, now); err != nil {
		return nil, fmt.Errorf("update login timestamp: %w", err)
	}

	user.LastLoginAt = &now
	user.PasswordHash = ""

	return &AuthResult{
		User:        user,
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.jwt.ttl.Seconds()),
	}, nil
}

func normalizeAuthEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validEmail(email string) bool {
	address, err := mail.ParseAddress(email)

	return err == nil && address.Address == email
}
