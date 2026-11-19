package users

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestPasswordServiceHashAndVerify(t *testing.T) {
	service := NewPasswordService()
	password := "SecurePassword123!"

	hash, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if hash == password {
		t.Fatal("password must not be stored as plaintext")
	}

	if err := service.VerifyPassword(password, hash); err != nil {
		t.Fatalf("password verification failed: %v", err)
	}
}

func TestPasswordServiceRejectsIncorrectPassword(t *testing.T) {
	service := NewPasswordService()

	hash, err := service.HashPassword("CorrectPassword123!")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	err = service.VerifyPassword("WrongPassword123!", hash)

	if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Fatalf("expected password mismatch, got: %v", err)
	}
}

func TestPasswordServiceRejectsEmptyPassword(t *testing.T) {
	service := NewPasswordService()

	_, err := service.HashPassword("")
	if !errors.Is(err, ErrEmptyPassword) {
		t.Fatalf("expected ErrEmptyPassword, got: %v", err)
	}

	err = service.VerifyPassword("", "some-hash")
	if !errors.Is(err, ErrEmptyPassword) {
		t.Fatalf("expected ErrEmptyPassword, got: %v", err)
	}
}

func TestPasswordServiceRejectsPasswordOver72Bytes(t *testing.T) {
	service := NewPasswordService()
	password := strings.Repeat("a", 73)

	_, err := service.HashPassword(password)
	if !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("expected ErrPasswordTooLong, got: %v", err)
	}

	err = service.VerifyPassword(password, "some-hash")
	if !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("expected ErrPasswordTooLong, got: %v", err)
	}
}

func TestPasswordServiceRejectsInvalidHash(t *testing.T) {
	service := NewPasswordService()

	err := service.VerifyPassword("Password123!", "invalid-hash")
	if err == nil {
		t.Fatal("expected invalid hash error")
	}
}
