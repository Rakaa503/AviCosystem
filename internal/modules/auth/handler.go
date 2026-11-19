package auth

import (
	"errors"
	"net/http"

	"avicosystem/backend/internal/modules/users"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service *AuthenticationService
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Email     string       `json:"email"`
	Role      users.Role   `json:"role"`
	Status    users.Status `json:"status"`
	CreatedAt interface{}  `json:"created_at"`
}

type loginResponse struct {
	User        userResponse `json:"user"`
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int64        `json:"expires_in"`
}

func NewHandler(service *AuthenticationService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Register(c fiber.Ctx) error {
	var request registerRequest

	if err := c.Bind().Body(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	user, err := h.service.Register(RegisterInput{
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
	})

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid registration input",
			})

		case errors.Is(err, ErrEmailAlreadyExists):
			return c.Status(http.StatusConflict).JSON(fiber.Map{
				"status":  "error",
				"message": "Email already registered",
			})

		default:
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Registration failed",
			})
		}
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"status": "success",
		"data":   toUserResponse(user),
	})
}

func (h *Handler) Login(c fiber.Ctx) error {
	var request loginRequest

	if err := c.Bind().Body(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	result, err := h.service.Login(LoginInput{
		Email:    request.Email,
		Password: request.Password,
	})

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid email or password",
			})

		case errors.Is(err, ErrAccountUnavailable):
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "Account is unavailable",
			})

		default:
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Login failed",
			})
		}
	}

	return c.Status(http.StatusOK).JSON(loginResponse{
		User:        toUserResponse(result.User),
		AccessToken: result.AccessToken,
		TokenType:   result.TokenType,
		ExpiresIn:   result.ExpiresIn,
	})
}

func toUserResponse(user *users.User) userResponse {
	return userResponse{
		ID:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
	}
}
