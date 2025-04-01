package domains

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TokenPermission string

const (
	PermissionRead  TokenPermission = "read"
	PermissionWrite TokenPermission = "write"
	PermissionAdmin TokenPermission = "admin"
)

type APIToken struct {
	ID          uuid.UUID         `json:"id" db:"id"`
	CompanyID   uuid.UUID         `json:"company_id" db:"company_id"`
	UserID      *uuid.UUID        `json:"user_id,omitempty" db:"user_id"`
	Token       string            `json:"token" db:"token"`
	Name        string            `json:"name" db:"name"`
	Permissions []TokenPermission `json:"permissions" db:"permissions"`
	LastUsedAt  *time.Time        `json:"last_used_at,omitempty" db:"last_used_at"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}

func (t *APIToken) SetPermissions(permissions []TokenPermission) error {
	jsonBytes, err := json.Marshal(permissions)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, &t.Permissions)
}

type APITokenRepository interface {
	Create(token *APIToken) error
	GetByID(id uuid.UUID) (*APIToken, error)
	GetByToken(token string) (*APIToken, error)
	GetByCompanyID(companyID uuid.UUID) ([]APIToken, error)
	GetByUserID(userID uuid.UUID) ([]APIToken, error)
	Update(token *APIToken) error
	Delete(id uuid.UUID) error
	List(companyID uuid.UUID, limit, offset int) ([]APIToken, error)
}

type APITokenService interface {
	Create(token *APIToken) error
	GetByID(id uuid.UUID) (*APIToken, error)
	GetByToken(token string) (*APIToken, error)
	GetByCompanyID(companyID uuid.UUID) ([]APIToken, error)
	GetByUserID(userID uuid.UUID) ([]APIToken, error)
	Update(token *APIToken) error
	Delete(id uuid.UUID) error
	List(companyID uuid.UUID, limit, offset int) ([]APIToken, error)
	ValidateToken(token string) (*APIToken, error)
	HasPermission(token *APIToken, permission TokenPermission) bool
}
