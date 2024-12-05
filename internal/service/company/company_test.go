package company

import (
	"context"
	"testing"
	"xm_test/internal/conf"
	"xm_test/internal/db"
	"xm_test/internal/db/options"
	"xm_test/internal/enum"
	"xm_test/internal/helpers"
	"xm_test/internal/mocks"
	"xm_test/internal/service/inputs"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/zap"
)

type companySuite struct {
	db db.DatabaseAdapter
	cs *company

	container *postgres.PostgresContainer
	suite.Suite
}

func (s *companySuite) SetupSuite() {
	conf.SetupConfig()
	conf.GlobalConfig.DatabaseType = enum.Postgres

	// establish connection to the test database with testcontainer
	user, database, password := "test", "test", "test"
	container, connStr, err := mocks.RunPostgresTestDatabaseContainer(
		user,
		database,
		password,
		conf.GlobalConfig.Postgres.InitScript,
	)
	s.Require().NoError(err)
	s.container = container

	mappedPort, err := container.MappedPort(context.Background(), nat.Port("5432"))
	s.Require().NoError(err)

	conf.GlobalConfig.Postgres.Port = mappedPort.Port()

	logger := zap.NewExample().Sugar()
	db := db.NewDatabaseAdapter(logger, options.WithConnectionString(*connStr))
	s.db = db
	s.cs = NewCompanyResolver(logger, db)
}

func (s *companySuite) TearDownSuite() {
	ctx := context.Background()
	s.Require().NoError(s.db.Close(ctx))
	s.Require().NoError(s.container.Terminate(ctx))
}

func (s *companySuite) TestCreateCompany() {
	s.Run("ok", func() {
		company := &inputs.CreateCompanyInput{
			Name:            "test",
			Description:     "test",
			AmountEmployees: new(int),
			Registered:      new(bool),
			Type:            enum.Corporation.String(),
		}
		storedCompany, err := s.cs.CreateCompany(company)
		s.Require().NoError(err)

		// check if company is created
		ctx := context.Background()
		companyModel, err := s.db.GetCompanyByID(ctx, storedCompany.ID.String())
		s.Require().NoError(err)

		s.Equal(company.Name, companyModel.Name)
		s.Equal(company.Description, companyModel.Description)
		s.Equal(*company.AmountEmployees, companyModel.AmountEmployees)
		s.Equal(*company.Registered, companyModel.Registered)
		s.Equal(company.Type, companyModel.Type)
	})
}

func (s *companySuite) TestGetCompanyByID() {
	// create a company
	company := &inputs.CreateCompanyInput{
		Name:            "TestComp",
		Description:     "test",
		AmountEmployees: helpers.PointerValue(10),
		Registered:      helpers.PointerValue(true),
		Type:            enum.Corporation.String(),
	}
	storedCompany, err := s.cs.CreateCompany(company)
	s.Require().NoError(err)

	s.Run("ok", func() {
		companyModel, err := s.cs.GetCompanyByID(storedCompany.ID.String())
		s.Require().NoError(err)

		s.Equal(storedCompany.ID, companyModel.ID)
		s.Equal(company.Name, companyModel.Name)
		s.Equal(company.Description, companyModel.Description)
		s.Equal(*company.AmountEmployees, companyModel.AmountEmployees)
		s.Equal(*company.Registered, companyModel.Registered)
		s.Equal(company.Type, companyModel.Type)
	})
}

func (s *companySuite) TestUpdateCompany() {
	// create a company
	company := &inputs.CreateCompanyInput{
		Name:            "update",
		Description:     "update",
		AmountEmployees: helpers.PointerValue(10),
		Registered:      helpers.PointerValue(true),
		Type:            enum.Corporation.String(),
	}
	storedCompany, err := s.cs.CreateCompany(company)
	s.Require().NoError(err)

	s.Run("ok", func() {
		updatedCompany := &inputs.UpdateCompany{
			Name:            "UpdatedComp",
			Description:     "updated",
			AmountEmployees: helpers.PointerValue(20),
			Registered:      helpers.PointerValue(false),
			Type:            enum.NonProfit.String(),
		}
		err := s.cs.UpdateCompany(storedCompany.ID.String(), updatedCompany)
		s.Require().NoError(err)

		// check if company has been updated
		companyModel, err := s.db.GetCompanyByID(context.Background(), storedCompany.ID.String())
		s.Require().NoError(err)

		s.Equal(updatedCompany.Name, companyModel.Name)
		s.Equal(updatedCompany.Description, companyModel.Description)
		s.Equal(*updatedCompany.AmountEmployees, companyModel.AmountEmployees)
		s.Equal(*updatedCompany.Registered, companyModel.Registered)
		s.Equal(updatedCompany.Type, companyModel.Type)
	})
}

func (s *companySuite) TestDeleteCompany() {
	// create a company
	company := &inputs.CreateCompanyInput{
		Name:            "delete",
		Description:     "delete",
		AmountEmployees: helpers.PointerValue(10),
		Registered:      helpers.PointerValue(true),
		Type:            enum.Corporation.String(),
	}
	storedCompany, err := s.cs.CreateCompany(company)
	s.Require().NoError(err)

	s.Run("ok", func() {
		err := s.cs.DeleteCompany(storedCompany.ID.String())
		s.Require().NoError(err)

		// check if company has been deleted
		_, err = s.db.GetCompanyByID(context.Background(), storedCompany.ID.String())
		s.Error(err)
	})
}

func TestCompanySuite(t *testing.T) {
	suite.Run(t, new(companySuite))
}
