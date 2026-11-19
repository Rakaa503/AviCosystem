package users

import (
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func setupUserRepository(t *testing.T) *Repository {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get SQL database: %v", err)
	}

	sqlDB.SetMaxOpenConns(1)

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("failed to migrate user model: %v", err)
	}

	return NewRepository(db)
}

func TestUserRepositoryCreateAndFind(t *testing.T) {
	repo := setupUserRepository(t)

	user := &User{
		Name:         "Rakha Avilla",
		Email:        "  RAKHA@EXAMPLE.COM ",
		PasswordHash: "hashed-password",
	}

	if err := repo.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if user.ID == uuid.Nil {
		t.Fatal("expected generated UUID")
	}

	if user.Email != "rakha@example.com" {
		t.Fatalf("unexpected normalized email: %s", user.Email)
	}

	if user.Role != RoleUser {
		t.Fatalf("unexpected role: %s", user.Role)
	}

	if user.Status != StatusActive {
		t.Fatalf("unexpected status: %s", user.Status)
	}

	found, err := repo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("failed to find user by ID: %v", err)
	}

	if found.Email != user.Email {
		t.Fatalf("unexpected email: %s", found.Email)
	}

	foundByEmail, err := repo.FindByEmail("RAKHA@example.com")
	if err != nil {
		t.Fatalf("failed to find user by email: %v", err)
	}

	if foundByEmail.ID != user.ID {
		t.Fatal("user IDs do not match")
	}
}

func TestUserRepositoryDuplicateEmail(t *testing.T) {
	repo := setupUserRepository(t)

	first := &User{
		Name:         "User One",
		Email:        "user@example.com",
		PasswordHash: "hash-one",
	}

	second := &User{
		Name:         "User Two",
		Email:        " USER@EXAMPLE.COM ",
		PasswordHash: "hash-two",
	}

	if err := repo.Create(first); err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}

	if err := repo.Create(second); err == nil {
		t.Fatal("expected duplicate email error")
	}
}

func TestUserRepositoryNotFound(t *testing.T) {
	repo := setupUserRepository(t)

	_, err := repo.FindByID(uuid.New())
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got: %v", err)
	}

	_, err = repo.FindByEmail("missing@example.com")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got: %v", err)
	}
}

func TestUserRepositoryDelete(t *testing.T) {
	repo := setupUserRepository(t)

	user := &User{
		Name:         "Delete User",
		Email:        "delete@example.com",
		PasswordHash: "hashed-password",
	}

	if err := repo.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if err := repo.Delete(user.ID); err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	_, err := repo.FindByID(user.ID)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected deleted user to be unavailable, got: %v", err)
	}
}
