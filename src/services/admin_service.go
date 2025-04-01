package services

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	db *sql.DB
}

type Company struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	APIKey            string    `json:"apiKey"`
	ActiveConnections int       `json:"activeConnections"`
	CreatedAt         time.Time `json:"createdAt"`
}

type Connection struct {
	ID        int       `json:"id"`
	CompanyID int       `json:"companyId"`
	UserID    string    `json:"userId"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type DashboardStats struct {
	TotalCompanies     int       `json:"totalCompanies"`
	ActiveConnections  int       `json:"activeConnections"`
	TotalMessagesToday int       `json:"totalMessages"`
	Companies          []Company `json:"companies"`
}

func NewAdminService(db *sql.DB) *AdminService {
	return &AdminService{db: db}
}

func (s *AdminService) Login(username, password string) (string, error) {
	var passwordHash string
	err := s.db.QueryRow("SELECT password_hash FROM master_admin WHERE username = $1", username).Scan(&passwordHash)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token here
	token := uuid.New().String() // This is a temporary solution, use proper JWT in production
	return token, nil
}

func (s *AdminService) CreateCompany(name string) (*Company, error) {
	apiKey := uuid.New().String()

	var company Company
	err := s.db.QueryRow(
		"INSERT INTO companies (name, api_key) VALUES ($1, $2) RETURNING id, name, api_key, created_at",
		name, apiKey,
	).Scan(&company.ID, &company.Name, &company.APIKey, &company.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &company, nil
}

func (s *AdminService) GetCompanies() ([]Company, error) {
	rows, err := s.db.Query(`
		SELECT c.id, c.name, c.api_key, c.created_at, 
		       COUNT(conn.id) as active_connections
		FROM companies c
		LEFT JOIN connections conn ON c.id = conn.company_id AND conn.status = 'connected'
		GROUP BY c.id, c.name, c.api_key, c.created_at
		ORDER BY c.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []Company
	for rows.Next() {
		var company Company
		err := rows.Scan(&company.ID, &company.Name, &company.APIKey, &company.CreatedAt, &company.ActiveConnections)
		if err != nil {
			return nil, err
		}
		companies = append(companies, company)
	}
	return companies, nil
}

func (s *AdminService) DeleteCompany(id int) error {
	_, err := s.db.Exec("DELETE FROM companies WHERE id = $1", id)
	return err
}

func (s *AdminService) GetDashboardStats() (*DashboardStats, error) {
	var stats DashboardStats

	// Get total companies
	err := s.db.QueryRow("SELECT COUNT(*) FROM companies").Scan(&stats.TotalCompanies)
	if err != nil {
		return nil, err
	}

	// Get active connections
	err = s.db.QueryRow("SELECT COUNT(*) FROM connections WHERE status = 'connected'").Scan(&stats.ActiveConnections)
	if err != nil {
		return nil, err
	}

	// Get total messages today
	err = s.db.QueryRow(`
		SELECT COUNT(*) 
		FROM messages 
		WHERE DATE(created_at) = CURRENT_DATE
	`).Scan(&stats.TotalMessagesToday)
	if err != nil {
		return nil, err
	}

	// Get companies with their active connections
	stats.Companies, err = s.GetCompanies()
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (s *AdminService) VerifyCompanyAccess(companyID string, token string) bool {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) 
		FROM companies c
		JOIN connections conn ON c.id = conn.company_id
		WHERE c.id = $1 AND conn.token = $2 AND conn.status = 'connected'
	`, companyID, token).Scan(&count)

	if err != nil {
		return false
	}
	return count > 0
}
