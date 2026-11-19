package users

import (
	"errors"

	"github.com/google/uuid"
)

var ErrUserUnavailable = errors.New("user account is unavailable")

type ProfileService struct {
	repository *Repository
}

func NewProfileService(repository *Repository) *ProfileService {
	return &ProfileService{
		repository: repository,
	}
}

func (s *ProfileService) GetProfile(userID uuid.UUID) (*User, error) {
	if userID == uuid.Nil {
		return nil, ErrUserNotFound
	}

	user, err := s.repository.FindByID(userID)
	if err != nil {
		return nil, err
	}

	if user.Status != StatusActive {
		return nil, ErrUserUnavailable
	}

	return user, nil
}
