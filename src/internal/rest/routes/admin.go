package routes

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/internal/rest/middleware"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/database/models"
	"github.com/gofiber/fiber/v2"
)

// SetupAdminRoutes sets up all admin routes
func SetupAdminRoutes(app *fiber.App) {
	admin := app.Group("/admin")
	admin.Use(middleware.AdminAuth())

	// Company management
	admin.Get("/companies", getCompanies)
	admin.Post("/companies", createCompany)
	admin.Get("/companies/:id", getCompany)
	admin.Put("/companies/:id", updateCompany)
	admin.Delete("/companies/:id", deleteCompany)

	// User management
	admin.Get("/users", getUsers)
	admin.Post("/users", createUser)
	admin.Get("/users/:id", getUser)
	admin.Put("/users/:id", updateUser)
	admin.Delete("/users/:id", deleteUser)
}

// Company handlers
func getCompanies(c *fiber.Ctx) error {
	companies, err := models.GetCompanies()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get companies",
		})
	}
	return c.JSON(companies)
}

func createCompany(c *fiber.Ctx) error {
	company := new(models.Company)
	if err := c.BodyParser(company); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := models.CreateCompany(company); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create company",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(company)
}

func getCompany(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid company ID",
		})
	}

	company, err := models.GetCompanyByID(int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Company not found",
		})
	}

	return c.JSON(company)
}

func updateCompany(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid company ID",
		})
	}

	company := new(models.Company)
	if err := c.BodyParser(company); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	company.ID = int64(id)
	if err := models.UpdateCompany(company); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update company",
		})
	}

	return c.JSON(company)
}

func deleteCompany(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid company ID",
		})
	}

	if err := models.DeleteCompany(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete company",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// User handlers
func getUsers(c *fiber.Ctx) error {
	users, err := models.GetUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get users",
		})
	}
	return c.JSON(users)
}

func createUser(c *fiber.Ctx) error {
	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := models.CreateUser(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

func getUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	user, err := models.GetUserByID(int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(user)
}

func updateUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	user.ID = int64(id)
	if err := models.UpdateUser(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user",
		})
	}

	return c.JSON(user)
}

func deleteUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	if err := models.DeleteUser(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete user",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
