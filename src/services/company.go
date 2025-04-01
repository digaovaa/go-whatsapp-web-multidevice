package services

import (
	"strconv"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/database/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CompanyService handles company-related operations
type CompanyService struct {
	db *gorm.DB
}

// NewCompanyService creates a new company service instance
func NewCompanyService(db *gorm.DB) *CompanyService {
	return &CompanyService{
		db: db,
	}
}

// Create creates a new company
func (s *CompanyService) Create(c *fiber.Ctx) error {
	var input struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	company := &models.Company{
		Name: input.Name,
	}

	if err := models.CreateCompany(company); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"company": company,
	})
}

// Get returns a company by ID
func (s *CompanyService) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	companyID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid company ID",
		})
	}

	company, err := models.GetCompanyByID(companyID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Company not found",
		})
	}

	return c.JSON(fiber.Map{
		"company": company,
	})
}

// Update updates a company
func (s *CompanyService) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	companyID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid company ID",
		})
	}

	company, err := models.GetCompanyByID(companyID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Company not found",
		})
	}

	var input struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	company.Name = input.Name

	if err := models.UpdateCompany(company); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"company": company,
	})
}

// Delete deletes a company
func (s *CompanyService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	companyID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid company ID",
		})
	}

	if err := models.DeleteCompany(companyID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Company deleted successfully",
	})
}
