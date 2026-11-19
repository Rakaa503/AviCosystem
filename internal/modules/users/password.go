package users

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmptyPassword   = errors.New("password cannot be empty")
	ErrPasswordTooLong = errors.New("password exceeds bcrypt 72-byte limit")
)

type PasswordService struct {
	cost int
}

func NewPasswordService() *PasswordService {
	return &PasswordService{
		cost: bcrypt.DefaultCost,
	}
}

func (s *PasswordService) HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	if len(password) > 72 {
		return "", ErrPasswordTooLong
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		s.cost,
	)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (s *PasswordService) VerifyPassword(
	password string,
	hashedPassword string,
) error {
	if password == "" {
		return ErrEmptyPassword
	}

	if len(password) > 72 {
		return ErrPasswordTooLong
	}

	return bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(password),
	)
}
