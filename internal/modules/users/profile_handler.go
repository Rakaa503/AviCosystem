package users

import (
	"errors"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ProfileHandler struct {
	service *ProfileService
}

type profileResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      Role      `json:"role"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func NewProfileHandler(service *ProfileService) *ProfileHandler {
	return &ProfileHandler{
		service: service,
	}
}

func (h *ProfileHandler) Me(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Authentication required",
		})
	}

	user, err := h.service.GetProfile(userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "User not found",
			})

		case errors.Is(err, ErrUserUnavailable):
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "Account is unavailable",
			})

		default:
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to retrieve profile",
			})
		}
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": profileResponse{
			ID:        user.ID.String(),
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
		},
	})
}
