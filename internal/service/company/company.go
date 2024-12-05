package company

import (
	"context"
	apierrors "xm_test/internal/api_errors"
	"xm_test/internal/db"
	"xm_test/internal/db/models"
	"xm_test/internal/service/inputs"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type company struct {
	logger *zap.SugaredLogger
	db     db.DatabaseAdapter
}

// NewCompanyResolver returns a new company service instance
func NewCompanyResolver(logger *zap.SugaredLogger, db db.DatabaseAdapter) *company {
	return &company{
		logger: logger,
		db:     db,
	}
}

// CreateCompany creates a new company
func (s *company) CreateCompany(company *inputs.CreateCompanyInput) (*models.CompanyModel, error) {
	s.logger.Infof("creating company with name '%s'", company.Name)
	ctx := context.Background()
	id := uuid.New()
	companyModel := models.CompanyModel{
		ID:              id,
		Name:            company.Name,
		Description:     company.Description,
		AmountEmployees: *company.AmountEmployees,
		Registered:      *company.Registered,
		Type:            company.Type,
	}
	if err := s.db.CreateCompany(ctx, &companyModel); err != nil {
		return nil, err
	}
	s.logger.Infof("company with name '%s' registered", company.Name)
	return &companyModel, nil
}

// GetCompanyByID retrieves a company by its ID
func (s *company) GetCompanyByID(id string) (*models.CompanyModel, error) {
	s.logger.Infof("retrieving company with id '%s'", id)

	s.logger.Debugf("checking uuid is valid")
	if _, err := uuid.Parse(id); err != nil {
		return nil, apierrors.ErrInvalidUUID
	}
	s.logger.Debugf("uuid is valid")

	ctx := context.Background()
	company, err := s.db.GetCompanyByID(ctx, id)
	if err != nil {
		return nil, err
	}
	s.logger.Infof("company with id '%s' retrieved", id)
	return company, nil
}

// UpdateCompany updates a company
func (s *company) UpdateCompany(id string, company *inputs.UpdateCompany) error {
	s.logger.Infof("updating company with id '%s'", id)

	s.logger.Debugf("checking uuid is valid")
	uuid, err := uuid.Parse(id)
	if err != nil {
		return apierrors.ErrInvalidUUID
	}
	s.logger.Debugf("uuid is valid")

	ctx := context.Background()
	companyModel := models.CompanyModel{
		ID:              uuid,
		Name:            company.Name,
		Description:     company.Description,
		AmountEmployees: *company.AmountEmployees,
		Registered:      *company.Registered,
		Type:            company.Type,
	}
	if err := s.db.UpdateCompany(ctx, id, &companyModel); err != nil {
		return err
	}
	s.logger.Infof("company with id '%s' updated", id)
	return nil
}

// DeleteCompany deletes a company
func (s *company) DeleteCompany(id string) error {
	s.logger.Infof("deleting company with id '%s'", id)

	s.logger.Debugf("checking uuid is valid")
	if _, err := uuid.Parse(id); err != nil {
		return apierrors.ErrInvalidUUID
	}
	s.logger.Debugf("uuid is valid")

	ctx := context.Background()
	if err := s.db.DeleteCompany(ctx, id); err != nil {
		return err
	}
	s.logger.Infof("company with id '%s' deleted", id)
	return nil
}
