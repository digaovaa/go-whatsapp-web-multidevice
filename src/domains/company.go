package domains

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CompanyRepository interface {
	Create(company *Company) error
	GetByID(id uuid.UUID) (*Company, error)
	GetByEmail(email string) (*Company, error)
	Update(company *Company) error
	Delete(id uuid.UUID) error
	List(limit, offset int) ([]Company, error)
}

type CompanyService interface {
	Create(company *Company) error
	GetByID(id uuid.UUID) (*Company, error)
	GetByEmail(email string) (*Company, error)
	Update(company *Company) error
	Delete(id uuid.UUID) error
	List(limit, offset int) ([]Company, error)
}
