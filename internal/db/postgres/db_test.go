package postgres

import (
	"context"
	"testing"
	"xm_test/internal/conf"
	"xm_test/internal/crypto"
	"xm_test/internal/db/models"
	"xm_test/internal/db/options"
	"xm_test/internal/enum"
	"xm_test/internal/mocks"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/zap"
)

type PostgresSuite struct {
	db        *postgresDB
	logger    *zap.SugaredLogger
	container *postgres.PostgresContainer

	suite.Suite
}

func (s *PostgresSuite) SetupSuite() {
	conf.SetupConfig()
	conf.GlobalConfig.DatabaseType = "postgres"

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
	s.logger = logger
	s.db = NewPostgresAdapter(logger)

	// connect to the database
	err = s.db.Connect(context.Background(), options.WithConnectionString(*connStr))
	s.Require().NoError(err)
}

func (s *PostgresSuite) TearDownSuite() {
	ctx := context.Background()
	s.Require().NoError(s.db.Close(ctx))
	s.Require().NoError(s.container.Terminate(ctx))
}

func (s *PostgresSuite) TestCreateUser() {
	ctx := context.Background()

	s.Run("ok", func() {
		user := models.UserModel{
			ID:          uuid.New(),
			Email:       "register@test.es",
			EncPassword: crypto.Md5Hash("test"),
		}
		err := s.db.CreateUser(ctx, &user)
		s.Require().NoError(err)

		// check that the user was created
		createdUser, err := s.db.GetUserByEmail(ctx, user.Email)
		s.Require().NoError(err)

		s.Equal(user.ID, createdUser.ID)
		s.Equal(user.Email, createdUser.Email)
		s.Equal(user.EncPassword, createdUser.EncPassword)
	})
}

func (s *PostgresSuite) TestGetUserByEmail() {
	ctx := context.Background()

	// create a user
	user := models.UserModel{
		ID:          uuid.New(),
		Email:       "get@test.es",
		EncPassword: crypto.Md5Hash("test"),
	}
	err := s.db.CreateUser(ctx, &user)
	s.Require().NoError(err)

	s.Run("ok", func() {
		// get the user
		createdUser, err := s.db.GetUserByEmail(ctx, user.Email)
		s.Require().NoError(err)

		s.Equal(user.ID, createdUser.ID)
		s.Equal(user.Email, createdUser.Email)
		s.Equal(user.EncPassword, createdUser.EncPassword)
	})
}

func (s *PostgresSuite) TestCreateCompany() {
	ctx := context.Background()

	s.Run("ok", func() {
		company := models.CompanyModel{
			ID:              uuid.New(),
			Name:            "createComp",
			Description:     "test",
			AmountEmployees: 10,
			Registered:      true,
			Type:            enum.Cooperative.String(),
		}
		err := s.db.CreateCompany(ctx, &company)
		s.Require().NoError(err)

		// check that the company was created
		createdCompany, err := s.db.GetCompanyByID(ctx, company.ID.String())
		s.Require().NoError(err)

		s.Equal(company.ID, createdCompany.ID)
		s.Equal(company.Name, createdCompany.Name)
		s.Equal(company.Description, createdCompany.Description)
		s.Equal(company.AmountEmployees, createdCompany.AmountEmployees)
		s.Equal(company.Registered, createdCompany.Registered)
		s.Equal(company.Type, createdCompany.Type)
	})
}

func (s *PostgresSuite) TestGetCompanyByID() {
	ctx := context.Background()

	// create a company
	company := models.CompanyModel{
		ID:              uuid.New(),
		Name:            "getComp",
		Description:     "test",
		AmountEmployees: 10,
		Registered:      true,
		Type:            enum.Cooperative.String(),
	}
	err := s.db.CreateCompany(ctx, &company)
	s.Require().NoError(err)

	s.Run("ok", func() {
		// get the company
		createdCompany, err := s.db.GetCompanyByID(ctx, company.ID.String())
		s.Require().NoError(err)

		s.Equal(company.ID, createdCompany.ID)
		s.Equal(company.Name, createdCompany.Name)
		s.Equal(company.Description, createdCompany.Description)
		s.Equal(company.AmountEmployees, createdCompany.AmountEmployees)
		s.Equal(company.Registered, createdCompany.Registered)
		s.Equal(company.Type, createdCompany.Type)
	})
}

func (s *PostgresSuite) TestUpdateCompany() {
	ctx := context.Background()

	// create a company
	company := models.CompanyModel{
		ID:              uuid.New(),
		Name:            "updateComp",
		Description:     "test",
		AmountEmployees: 10,
		Registered:      true,
		Type:            enum.Cooperative.String(),
	}
	err := s.db.CreateCompany(ctx, &company)
	s.Require().NoError(err)

	s.Run("ok", func() {
		company.Name = "updatedComp"
		company.Description = "updated"
		company.AmountEmployees = 20
		company.Registered = false
		company.Type = enum.NonProfit.String()

		err := s.db.UpdateCompany(ctx, company.ID.String(), &company)
		s.Require().NoError(err)

		// check that the company was updated
		updatedCompany, err := s.db.GetCompanyByID(ctx, company.ID.String())
		s.Require().NoError(err)

		s.Equal(company.ID, updatedCompany.ID)
		s.Equal(company.Name, updatedCompany.Name)
		s.Equal(company.Description, updatedCompany.Description)
		s.Equal(company.AmountEmployees, updatedCompany.AmountEmployees)
		s.Equal(company.Registered, updatedCompany.Registered)
	})
}

func (s *PostgresSuite) TestDeleteCompany() {
	ctx := context.Background()

	// create a company
	company := models.CompanyModel{
		ID:              uuid.New(),
		Name:            "deleteComp",
		Description:     "test",
		AmountEmployees: 10,
		Registered:      true,
		Type:            enum.Cooperative.String(),
	}
	err := s.db.CreateCompany(ctx, &company)
	s.Require().NoError(err)

	s.Run("ok", func() {
		err := s.db.DeleteCompany(ctx, company.ID.String())
		s.Require().NoError(err)

		// check that the company was deleted
		_, err = s.db.GetCompanyByID(ctx, company.ID.String())
		s.Error(err)
	})
}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, new(PostgresSuite))
}
