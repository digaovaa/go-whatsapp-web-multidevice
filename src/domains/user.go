package domains

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	CompanyID    uuid.UUID `json:"company_id" db:"company_id"`
	Name         string    `json:"name" db:"name"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type UserRepository interface {
	Create(user *User) error
	GetByID(id uuid.UUID) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByCompanyID(companyID uuid.UUID) ([]User, error)
	Update(user *User) error
	Delete(id uuid.UUID) error
	List(companyID uuid.UUID, limit, offset int) ([]User, error)
}

type UserService interface {
	Create(user *User) error
	GetByID(id uuid.UUID) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByCompanyID(companyID uuid.UUID) ([]User, error)
	Update(user *User) error
	Delete(id uuid.UUID) error
	List(companyID uuid.UUID, limit, offset int) ([]User, error)
	ValidatePassword(user *User, password string) bool
}
