package rest

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/internal/rest/middleware"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/services"
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes initializes all API routes
func SetupRoutes(app *fiber.App, services *services.Services) {
	// Company routes
	company := app.Group("/api/v1/companies")
	company.Post("/", services.Company.Create)
	company.Get("/:id", services.Company.Get)
	company.Put("/:id", services.Company.Update)
	company.Delete("/:id", services.Company.Delete)

	// Company authenticated routes
	companyAuth := app.Group("/api/v1/company")
	companyAuth.Use(middleware.CompanyAuth())

	// User management routes
	companyAuth.Get("/users", services.User.List)
	companyAuth.Post("/users", services.User.Create)
	companyAuth.Get("/users/:id", services.User.Get)
	companyAuth.Put("/users/:id", services.User.Update)
	companyAuth.Delete("/users/:id", services.User.Delete)

	// WhatsApp routes
	whatsapp := app.Group("/api/v1/whatsapp/:userID")
	whatsapp.Use(middleware.UserAuth())

	whatsapp.Get("/status", services.WhatsApp.GetStatus)
	whatsapp.Post("/login", services.WhatsApp.Login)
	whatsapp.Post("/logout", services.WhatsApp.Logout)
	whatsapp.Post("/send", services.WhatsApp.SendMessage)
	whatsapp.Get("/groups", services.WhatsApp.ListGroups)
	whatsapp.Get("/contacts", services.WhatsApp.ListContacts)
}
