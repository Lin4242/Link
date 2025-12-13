package middleware

import (
	"strings"

	"link/internal/domain"
	"link/internal/pkg/token"

	"github.com/gofiber/fiber/v2"
)

func Auth(tm *token.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(401).JSON(fiber.Map{
				"error": fiber.Map{"code": domain.ErrCodeUnauthorized, "message": "missing token"},
			})
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims, err := tm.Verify(tokenStr)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": fiber.Map{"code": domain.ErrCodeUnauthorized, "message": err.Error()},
			})
		}

		c.Locals("userID", claims.UserID)
		return c.Next()
	}
}
