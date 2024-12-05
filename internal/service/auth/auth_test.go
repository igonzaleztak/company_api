package auth

import (
	"context"
	"testing"
	"xm_test/internal/conf"
	"xm_test/internal/crypto"
	"xm_test/internal/db"
	"xm_test/internal/db/options"
	"xm_test/internal/enum"
	"xm_test/internal/mocks"
	"xm_test/internal/token"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/zap"
)

type authSuite struct {
	db db.DatabaseAdapter
	as *auth

	container *postgres.PostgresContainer
	suite.Suite
}

func (s *authSuite) SetupSuite() {
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
	s.as = NewAuthResolver(logger, db)
}

func (s *authSuite) TearDownSuite() {
	ctx := context.Background()
	s.Require().NoError(s.db.Close(ctx))
	s.Require().NoError(s.container.Terminate(ctx))
}

func (s *authSuite) TestRegisterUser() {
	ctx := context.Background()
	_ = ctx

	s.Run("ok", func() {
		email := "testRegister@test.es"
		password := "password"

		err := s.as.Register(email, password)
		s.Require().NoError(err)

		user, err := s.db.GetUserByEmail(ctx, email)
		s.Require().NoError(err)

		s.Equal(email, user.Email)
		s.NotEmpty(user.EncPassword)

		encPwd := crypto.Md5Hash(password)
		s.Equal(encPwd, user.EncPassword)
	})
}

func (s *authSuite) TestLoginUser() {
	// register user in database
	email := "testLogin@test.es"
	password := "password"

	err := s.as.Register(email, password)
	s.Require().NoError(err)

	s.Run("ok", func() {
		accessToken, err := s.as.Login(email, password)
		s.Require().NoError(err)

		s.NotEmpty(accessToken)
		claims, err := token.ValidateAndParseToken(*accessToken)
		s.Require().NoError(err)

		s.Equal(claims.Email, email)
	})

	s.Run("invalid password", func() {
		invalidPassword := "invalid"
		_, err := s.as.Login(email, invalidPassword)
		s.Error(err)
	})
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(authSuite))
}
