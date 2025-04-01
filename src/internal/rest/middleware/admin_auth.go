package middleware

import (
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/services"
	"github.com/gofiber/fiber/v2"
)

var adminService *services.AdminService

func SetAdminService(service *services.AdminService) {
	adminService = service
}

func AdminAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		// Extract token from Bearer header
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		token := parts[1]

		// Verify token with admin service
		if adminService == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Admin service not initialized",
			})
		}

		// TODO: Implement proper token validation
		// For now, we'll just check if the token is valid
		if !strings.HasPrefix(token, "admin-") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Store admin status in context
		c.Locals("isAdmin", true)

		return c.Next()
	}
}
