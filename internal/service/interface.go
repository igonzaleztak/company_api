package service

import (
	"xm_test/internal/db"
	"xm_test/internal/db/models"
	"xm_test/internal/service/auth"
	"xm_test/internal/service/company"
	"xm_test/internal/service/inputs"

	"go.uber.org/zap"
)

// AuthService is an interface for the authentication service. It defines the Login, and Logout methods.
type AuthService interface {
	Register(email string, password string) error         // Register registers a new user
	Login(email string, password string) (*string, error) // Login logs in a user
}

// CompanyService is an interface for the company service.
type CompanyService interface {
	CreateCompany(company *inputs.CreateCompanyInput) (*models.CompanyModel, error) // CreateCompany creates a new company
	GetCompanyByID(id string) (*models.CompanyModel, error)                         // GetCompany retrieves a company by its ID
	UpdateCompany(id string, updatedCompany *inputs.UpdateCompany) error            // UpdateCompany updates a company by its ID
	DeleteCompany(id string) error                                                  // DeleteCompany deletes a company by its ID
}

// NewAuthService returns a new auth service instance
func NewAuthService(logger *zap.SugaredLogger, db db.DatabaseAdapter) AuthService {
	return auth.NewAuthResolver(logger, db)
}

// NewCompanyService returns a new company service instance
func NewCompanyService(logger *zap.SugaredLogger, db db.DatabaseAdapter) CompanyService {
	return company.NewCompanyResolver(logger, db)
}
