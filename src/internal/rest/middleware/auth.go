package middleware

import (
	"strconv"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/database/models"
	"github.com/gofiber/fiber/v2"
)

// Auth middleware checks for user token authentication
func Auth() fiber.Handler {
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

		// Get company by token
		company, err := models.GetCompanyByToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Store company info in context for later use
		c.Locals("company", company)
		return c.Next()
	}
}

// CompanyAuth middleware checks for valid company token
func CompanyAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		token := parts[1]
		companyID, err := models.GetCompanyIDByToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		c.Locals("companyID", companyID)
		return c.Next()
	}
}

// UserAuth middleware checks if the request has access to the specified user
func UserAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		companyID, ok := c.Locals("companyID").(int64)
		if !ok || companyID == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Company not authenticated",
			})
		}

		userID := c.Params("userID")
		if userID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing user ID",
			})
		}

		userIDInt, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid user ID format",
			})
		}

		user, err := models.GetUserByID(userIDInt)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		if user.CompanyID != companyID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		c.Locals("userID", user.ID)
		return c.Next()
	}
}
