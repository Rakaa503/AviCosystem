package users

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusInactive  Status = "inactive"
	StatusSuspended Status = "suspended"
)

type User struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name            string    `gorm:"size:100;not null"`
	Email           string    `gorm:"size:255;not null;uniqueIndex"`
	PasswordHash    string    `gorm:"size:255;not null"`
	Role            Role      `gorm:"size:20;not null;default:user"`
	Status          Status    `gorm:"size:20;not null;default:active;index"`
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	return nil
}

func (u *User) BeforeSave(tx *gorm.DB) error {
	u.Email = normalizeEmail(u.Email)

	if u.Role == "" {
		u.Role = RoleUser
	}

	if u.Status == "" {
		u.Status = StatusActive
	}

	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
