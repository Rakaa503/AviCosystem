
package users

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(user *User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	user.Name = strings.TrimSpace(user.Name)
	user.Email = normalizeEmail(user.Email)

	if user.Name == "" || user.Email == "" || user.PasswordHash == "" {
		return errors.New("name, email and password hash are required")
	}

	if len(user.Name) > 100 || len(user.Email) > 255 {
		return errors.New("user name or email exceeds maximum length")
	}

	if user.Role == "" {
		user.Role = RoleUser
	}

	if user.Status == "" {
		user.Status = StatusActive
	}

	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *Repository) FindByID(id uuid.UUID) (*User, error) {
	if id == uuid.Nil {
		return nil, ErrUserNotFound
	}

	var user User

	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("find user by ID: %w", err)
	}

	return &user, nil
}

func (r *Repository) FindByEmail(email string) (*User, error) {
	email = normalizeEmail(email)

	if email == "" {
		return nil, ErrUserNotFound
	}

	var user User

	if err := r.db.
		Where("email = ?", email).
		First(&user).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return &user, nil
}

func (r *Repository) UpdateLastLogin(
	id uuid.UUID,
	loginAt time.Time,
) error {
	if id == uuid.Nil {
		return ErrUserNotFound
	}

	result := r.db.
		Model(&User{}).
		Where("id = ?", id).
		Update("last_login_at", loginAt)

	if result.Error != nil {
		return fmt.Errorf("update last login: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *Repository) Delete(id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrUserNotFound
	}

	result := r.db.Delete(&User{}, "id = ?", id)

	if result.Error != nil {
		return fmt.Errorf("delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}
