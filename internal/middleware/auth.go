package middleware

import (
	"strings"

	authmodule "avicosystem/backend/internal/modules/auth"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const (
	LocalUserID   = "user_id"
	LocalUserRole = "user_role"
)

func Auth(jwtService *authmodule.JWTService) fiber.Handler {
	return func(c fiber.Ctx) error {
		authorization := strings.TrimSpace(c.Get("Authorization"))
		parts := strings.Fields(authorization)

		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return fiber.NewError(
				fiber.StatusUnauthorized,
				"Missing or invalid authorization header",
			)
		}

		claims, err := jwtService.ValidateAccessToken(parts[1])
		if err != nil {
			return fiber.NewError(
				fiber.StatusUnauthorized,
				"Invalid or expired access token",
			)
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil || userID == uuid.Nil {
			return fiber.NewError(
				fiber.StatusUnauthorized,
				"Invalid access token subject",
			)
		}

		c.Locals(LocalUserID, userID)
		c.Locals(LocalUserRole, claims.Role)

		return c.Next()
	}
}
