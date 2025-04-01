package domains

import (
	"time"

	"github.com/google/uuid"
)

type WhatsAppConnection struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	PhoneNumber     string     `json:"phone_number" db:"phone_number"`
	Status          string     `json:"status" db:"status"`
	QRCode          string     `json:"qr_code,omitempty" db:"qr_code"`
	LastConnectedAt *time.Time `json:"last_connected_at,omitempty" db:"last_connected_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type WhatsAppConnectionRepository interface {
	Create(conn *WhatsAppConnection) error
	GetByID(id uuid.UUID) (*WhatsAppConnection, error)
	GetByUserID(userID uuid.UUID) ([]WhatsAppConnection, error)
	GetByPhoneNumber(phoneNumber string) (*WhatsAppConnection, error)
	Update(conn *WhatsAppConnection) error
	Delete(id uuid.UUID) error
	List(userID uuid.UUID, limit, offset int) ([]WhatsAppConnection, error)
}

type WhatsAppConnectionService interface {
	Create(conn *WhatsAppConnection) error
	GetByID(id uuid.UUID) (*WhatsAppConnection, error)
	GetByUserID(userID uuid.UUID) ([]WhatsAppConnection, error)
	GetByPhoneNumber(phoneNumber string) (*WhatsAppConnection, error)
	Update(conn *WhatsAppConnection) error
	Delete(id uuid.UUID) error
	List(userID uuid.UUID, limit, offset int) ([]WhatsAppConnection, error)
	Connect(id uuid.UUID) error
	Disconnect(id uuid.UUID) error
	GetQRCode(id uuid.UUID) (string, error)
}
