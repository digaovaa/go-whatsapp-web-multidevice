package services

import (
	"gorm.io/gorm"
)

// Services holds all service instances
type Services struct {
	Company  *CompanyService
	User     *UserService
	WhatsApp *WhatsAppService
}

// NewServices creates a new Services instance
func NewServices(db *gorm.DB) *Services {
	return &Services{
		Company:  NewCompanyService(db),
		User:     NewUserServiceDB(db),
		WhatsApp: NewWhatsAppService(db),
	}
}
