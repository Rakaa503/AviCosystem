package users

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestProfileServiceGetProfile(t *testing.T) {
	repository := setupUserRepository(t)

	user := &User{
		Name:         "Rakha Avilla",
		Email:        "rakha@example.com",
		PasswordHash: "hashed-password",
		Status:       StatusActive,
	}

	if err := repository.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	service := NewProfileService(repository)

	found, err := service.GetProfile(user.ID)
	if err != nil {
		t.Fatalf("failed to get profile: %v", err)
	}

	if found.ID != user.ID {
		t.Fatalf("unexpected user ID: %s", found.ID)
	}

	if found.Email != user.Email {
		t.Fatalf("unexpected email: %s", found.Email)
	}
}

func TestProfileServiceRejectsUnavailableAccount(t *testing.T) {
	repository := setupUserRepository(t)

	user := &User{
		Name:         "Inactive User",
		Email:        "inactive@example.com",
		PasswordHash: "hashed-password",
		Status:       StatusInactive,
	}

	if err := repository.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	service := NewProfileService(repository)

	_, err := service.GetProfile(user.ID)
	if !errors.Is(err, ErrUserUnavailable) {
		t.Fatalf("expected ErrUserUnavailable, got: %v", err)
	}
}

func TestProfileServiceRejectsMissingUser(t *testing.T) {
	repository := setupUserRepository(t)
	service := NewProfileService(repository)

	_, err := service.GetProfile(uuid.New())
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got: %v", err)
	}
}

func TestProfileServiceRejectsNilUUID(t *testing.T) {
	repository := setupUserRepository(t)
	service := NewProfileService(repository)

	_, err := service.GetProfile(uuid.Nil)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got: %v", err)
	}
}
