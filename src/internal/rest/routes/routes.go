package routes

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/internal/rest/middleware"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/database"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/services"
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes sets up all routes
func SetupRoutes(app *fiber.App) {
	// Initialize services
	svc := services.NewServices(database.DB)

	// Public routes
	app.Post("/login", login)
	app.Post("/register", register)

	// Protected routes
	api := app.Group("/api")
	api.Use(middleware.Auth())

	// WhatsApp routes
	whatsapp := api.Group("/whatsapp")
	whatsapp.Get("/status", svc.WhatsApp.GetStatus)
	whatsapp.Post("/login", svc.WhatsApp.Login)
	whatsapp.Post("/logout", svc.WhatsApp.Logout)
	whatsapp.Post("/send", svc.WhatsApp.SendMessage)
	whatsapp.Get("/groups", svc.WhatsApp.ListGroups)
	whatsapp.Get("/contacts", svc.WhatsApp.ListContacts)
}

// Login handler
func login(c *fiber.Ctx) error {
	// TODO: Implement login logic
	return c.SendStatus(fiber.StatusNotImplemented)
}

// Register handler
func register(c *fiber.Ctx) error {
	// TODO: Implement register logic
	return c.SendStatus(fiber.StatusNotImplemented)
}
