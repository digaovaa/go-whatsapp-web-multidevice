package middleware

import (
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/database/models"
	"github.com/gofiber/fiber/v2"
)

// AdminAuth middleware checks for master token authentication
func AdminAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token format",
			})
		}

		admin, err := models.GetAdminByToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Store admin info in context for later use
		c.Locals("admin", admin)
		return c.Next()
	}
}
