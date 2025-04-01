package rest

import (
	"strconv"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/services"
	"github.com/gofiber/fiber/v2"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateCompanyRequest struct {
	Name string `json:"name"`
}

func InitRestAdmin(app *fiber.App, adminService *services.AdminService) {
	admin := app.Group("/api/admin")

	// Admin login
	admin.Post("/login", func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		token, err := adminService.Login(req.Username, req.Password)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		return c.JSON(fiber.Map{
			"token": token,
		})
	})

	// Get dashboard stats
	admin.Get("/dashboard", func(c *fiber.Ctx) error {
		stats, err := adminService.GetDashboardStats()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get dashboard stats",
			})
		}

		return c.JSON(stats)
	})

	// Get all companies
	admin.Get("/companies", func(c *fiber.Ctx) error {
		companies, err := adminService.GetCompanies()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get companies",
			})
		}

		return c.JSON(companies)
	})

	// Create new company
	admin.Post("/companies", func(c *fiber.Ctx) error {
		var req CreateCompanyRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		company, err := adminService.CreateCompany(req.Name)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create company",
			})
		}

		return c.JSON(company)
	})

	// Delete company
	admin.Delete("/companies/:id", func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid company ID",
			})
		}

		if err := adminService.DeleteCompany(id); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete company",
			})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})
}
