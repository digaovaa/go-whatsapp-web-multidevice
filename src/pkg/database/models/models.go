package models

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"gorm.io/gorm"
)

var DB *gorm.DB

// SetDB sets the database connection
func SetDB(db *gorm.DB) {
	DB = db
}

// Admin represents the system administrator
type Admin struct {
	ID          int64  `gorm:"primaryKey"`
	Username    string `gorm:"unique;not null"`
	Password    string `gorm:"not null"`
	MasterToken string `gorm:"unique;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Company represents a company in the system
type Company struct {
	ID        int64  `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Token     string `gorm:"unique;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// User represents a user in the system
type User struct {
	ID        int64  `gorm:"primaryKey"`
	CompanyID int64  `gorm:"not null"`
	Name      string `gorm:"not null"`
	Email     string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	Active    bool   `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateAdmin creates a new admin user
func CreateAdmin(admin *Admin) error {
	token, err := generateToken()
	if err != nil {
		return err
	}
	admin.MasterToken = token
	return DB.Create(admin).Error
}

// GetAdminByToken returns an admin by master token
func GetAdminByToken(token string) (*Admin, error) {
	var admin Admin
	result := DB.Where("master_token = ?", token).First(&admin)
	return &admin, result.Error
}

// CreateCompany creates a new company
func CreateCompany(company *Company) error {
	token, err := generateToken()
	if err != nil {
		return err
	}
	company.Token = token
	return DB.Create(company).Error
}

// GetCompanyByID returns a company by ID
func GetCompanyByID(id int64) (*Company, error) {
	var company Company
	result := DB.First(&company, id)
	return &company, result.Error
}

// GetCompanyIDByToken returns the company ID for a given token
func GetCompanyIDByToken(token string) (int64, error) {
	var company Company
	result := DB.Where("token = ?", token).First(&company)
	if result.Error != nil {
		return 0, result.Error
	}
	return company.ID, nil
}

// UpdateCompany updates an existing company
func UpdateCompany(company *Company) error {
	return DB.Save(company).Error
}

// DeleteCompany deletes a company
func DeleteCompany(id int64) error {
	return DB.Delete(&Company{}, id).Error
}

// GetActiveUsers returns all active users for a company
func GetActiveUsers(companyID int64) ([]User, error) {
	var users []User
	result := DB.Where("company_id = ? AND active = ?", companyID, true).Find(&users)
	return users, result.Error
}

// GetUsers returns all users
func GetUsers() ([]User, error) {
	var users []User
	result := DB.Find(&users)
	return users, result.Error
}

// GetUserByID returns a user by ID
func GetUserByID(id int64) (*User, error) {
	var user User
	result := DB.First(&user, id)
	return &user, result.Error
}

// CreateUser creates a new user
func CreateUser(user *User) error {
	return DB.Create(user).Error
}

// UpdateUser updates a user
func UpdateUser(user *User) error {
	return DB.Save(user).Error
}

// DeleteUser deletes a user by ID
func DeleteUser(id int64) error {
	return DB.Delete(&User{}, id).Error
}

// generateToken generates a random token for company authentication
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetCompanies retrieves all companies from the database
func GetCompanies() ([]Company, error) {
	var companies []Company
	if err := DB.Find(&companies).Error; err != nil {
		return nil, err
	}
	return companies, nil
}

// GetCompanyByToken returns a company by token
func GetCompanyByToken(token string) (*Company, error) {
	var company Company
	result := DB.Where("token = ?", token).First(&company)
	return &company, result.Error
}
