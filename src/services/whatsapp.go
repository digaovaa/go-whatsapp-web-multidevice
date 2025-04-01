package services

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/whatsapp"
	"github.com/gofiber/fiber/v2"
	"go.mau.fi/whatsmeow/binary/proto"
	"gorm.io/gorm"
)

// WhatsAppService handles WhatsApp-related operations
type WhatsAppService struct {
	db *gorm.DB
}

// NewWhatsAppService creates a new WhatsApp service instance
func NewWhatsAppService(db *gorm.DB) *WhatsAppService {
	return &WhatsAppService{
		db: db,
	}
}

// GetStatus returns the connection status for a user
func (s *WhatsAppService) GetStatus(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	instance := whatsapp.GetInstanceManager().GetInstance(userID)
	if instance == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "WhatsApp instance not found",
		})
	}

	return c.JSON(fiber.Map{
		"status": instance.Client.Store.ID == nil,
	})
}

// Login initiates WhatsApp login for a user
func (s *WhatsAppService) Login(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	instance := whatsapp.GetInstanceManager().GetInstance(userID)
	if instance == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "WhatsApp instance not found",
		})
	}

	qrChan, _ := instance.Client.GetQRChannel(c.Context())
	err := instance.Client.Connect()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	evt := <-qrChan
	if evt.Event == "code" {
		return c.JSON(fiber.Map{
			"qr": evt.Code,
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "Failed to get QR code",
	})
}

// Logout logs out a user from WhatsApp
func (s *WhatsAppService) Logout(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	instance := whatsapp.GetInstanceManager().GetInstance(userID)
	if instance == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "WhatsApp instance not found",
		})
	}

	instance.Client.Disconnect()
	whatsapp.GetInstanceManager().RemoveInstance(userID)

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// SendMessage sends a WhatsApp message
func (s *WhatsAppService) SendMessage(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	instance := whatsapp.GetInstanceManager().GetInstance(userID)
	if instance == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "WhatsApp instance not found",
		})
	}

	var input struct {
		To      string `json:"to"`
		Message string `json:"message"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	recipient, err := whatsapp.ParseJID(input.To)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	msg := &proto.Message{
		Conversation: &input.Message,
	}

	_, err = instance.Client.SendMessage(c.Context(), recipient, msg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Message sent successfully",
	})
}

// ListGroups returns the list of WhatsApp groups
func (s *WhatsAppService) ListGroups(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	instance := whatsapp.GetInstanceManager().GetInstance(userID)
	if instance == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "WhatsApp instance not found",
		})
	}

	groups, err := instance.Client.GetJoinedGroups()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"groups": groups,
	})
}

// ListContacts returns the list of WhatsApp contacts
func (s *WhatsAppService) ListContacts(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	instance := whatsapp.GetInstanceManager().GetInstance(userID)
	if instance == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "WhatsApp instance not found",
		})
	}

	contacts, err := instance.Client.Store.Contacts.GetAllContacts()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"contacts": contacts,
	})
}
