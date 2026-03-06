package httpserver

import (
	"crypto/subtle"

	"github.com/gofiber/fiber/v2"
)

// APIKeyMiddleware authorizes requests using X-API-Key.
func APIKeyMiddleware(expected string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		provided := c.Get("X-API-Key")
		if provided == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing api key"})
		}
		if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "invalid api key"})
		}
		return c.Next()
	}
}
