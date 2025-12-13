package middleware

import "github.com/gofiber/fiber/v2"

func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Force no caching
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		return c.Next()
	}
}
