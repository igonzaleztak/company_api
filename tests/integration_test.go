package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
	apierrors "xm_test/internal/api_errors"
	"xm_test/internal/conf"
	"xm_test/internal/db/models"
	"xm_test/internal/enum"
	"xm_test/internal/helpers"
	"xm_test/internal/mocks"
	"xm_test/internal/transport/http/schemas"

	"github.com/docker/go-connections/nat"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type integrationSuite struct {
	apiURL string

	email   string // email for test user account created in the setup
	pasword string // password for test user account created in the setup

	apiContainer *testcontainers.Container
	pgContainer  *postgres.PostgresContainer
	suite.Suite
}

func (s *integrationSuite) SetupSuite() {
	err := godotenv.Load(".env.docker")
	s.Require().NoError(err)

	conf.SetupConfig()

	// launch database
	pgContainer, _, err := mocks.RunPostgresTestDatabaseContainer(
		conf.GlobalConfig.Postgres.User,
		conf.GlobalConfig.Postgres.Database,
		conf.GlobalConfig.Postgres.Password,
		conf.GlobalConfig.Postgres.InitScript,
	)
	s.Require().NoError(err)
	s.pgContainer = pgContainer

	containerIP, err := pgContainer.ContainerIP(context.Background())
	s.Require().NoError(err)

	// launch api
	exposedPorts := []string{fmt.Sprintf("%s/tcp", conf.GlobalConfig.Port), fmt.Sprintf("%s/tcp", conf.GlobalConfig.HealthPort)}
	natHealthPort, err := nat.NewPort("tcp", conf.GlobalConfig.HealthPort)
	s.Require().NoError(err)

	env := map[string]string{
		"POSTGRES_HOST":     containerIP,
		"POSTGRES_PORT":     conf.GlobalConfig.Postgres.Port,
		"POSTGRES_USER":     conf.GlobalConfig.Postgres.User,
		"POSTGRES_PASSWORD": conf.GlobalConfig.Postgres.Password,
		"POSTGRES_DB":       conf.GlobalConfig.Postgres.Database,
		"JWT_SECRET":        conf.GlobalConfig.JwtSecret,
		"PORT":              conf.GlobalConfig.Port,
		"HEALTH_PORT":       conf.GlobalConfig.HealthPort,
		"LOG_LEVEL":         conf.GlobalConfig.LogLevel.String(),
		"DATABASE_TYPE":     conf.GlobalConfig.DatabaseType.String(),
	}
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "../",
			Dockerfile: "Dockerfile",
		},
		Name:         "integration_test_api",
		ExposedPorts: exposedPorts,
		Env:          env,
		WaitingFor:   wait.ForHTTP("/health").WithPort(natHealthPort),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.Require().NoError(err)

	s.apiContainer = &container
	mappedPort, err := container.MappedPort(ctx, nat.Port(conf.GlobalConfig.Port))
	s.Require().NoError(err)
	s.apiURL = fmt.Sprintf("http://localhost:%s", mappedPort.Port())

	// create test user account
	s.email = "integration@test.com"
	s.pasword = "integration"

	body := `{"email":"` + s.email + `","password":"` + s.pasword + `"}`
	resp, err := http.Post(s.apiURL+"/register", "application/json", strings.NewReader(body))
	s.Require().NoError(err)

	s.Require().Equal(http.StatusCreated, resp.StatusCode)
}

func (s *integrationSuite) TearDownSuite() {
	err := testcontainers.TerminateContainer(*s.apiContainer)
	s.Require().NoError(err)

	err = s.pgContainer.Terminate(context.Background())
	s.Require().NoError(err)
}

func (s *integrationSuite) TestCreateCompany() {
	s.Run("ok", func() {
		// login
		bodyLogin := `{"email":"` + s.email + `","password":"` + s.pasword + `"}`
		loginResp, err := http.Post(s.apiURL+"/login", "application/json", strings.NewReader(bodyLogin))
		s.Require().NoError(err)
		s.Require().Equal(http.StatusOK, loginResp.StatusCode)

		var loginResponse schemas.LoginResponse
		err = json.NewDecoder(loginResp.Body).Decode(&loginResponse)
		s.Require().NoError(err)
		loginResp.Body.Close()

		input := []struct {
			in                *schemas.CreateCompanyRequest
			out               models.CompanyModel
			err               error
			exectedStatusCode int
		}{
			{
				in: &schemas.CreateCompanyRequest{
					Name:            "test",
					Description:     "test",
					AmountEmployees: helpers.PointerValue(10),
					Registered:      helpers.PointerValue(true),
					Type:            enum.Corporation.String(),
				},
				out: models.CompanyModel{
					Name:            "test",
					Description:     "test",
					AmountEmployees: 10,
					Registered:      true,
					Type:            enum.Corporation.String(),
				},
				err:               nil,
				exectedStatusCode: http.StatusOK,
			},
			{
				in: &schemas.CreateCompanyRequest{
					Name:            "test2",
					Description:     "test2",
					AmountEmployees: helpers.PointerValue(20),
					Registered:      helpers.PointerValue(false),
					Type:            enum.Cooperative.String(),
				},
				out: models.CompanyModel{
					Name:            "test2",
					Description:     "test2",
					AmountEmployees: 20,
					Registered:      false,
					Type:            enum.Cooperative.String(),
				},
				err:               nil,
				exectedStatusCode: http.StatusOK,
			},
		}

		for _, tt := range input {
			jsonData, err := json.Marshal(tt.in)
			s.Require().NoError(err)

			client := &http.Client{}
			req, err := http.NewRequest("POST", s.apiURL+"/company/create", bytes.NewBuffer(jsonData))
			s.Require().NoError(err)

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)

			resp, err := client.Do(req)
			s.Require().NoError(err)
			s.Equal(tt.exectedStatusCode, resp.StatusCode)

			defer resp.Body.Close()

			var response models.CompanyModel
			err = json.NewDecoder(resp.Body).Decode(&response)
			s.Require().NoError(err)

			s.Equal(tt.out.Name, response.Name)
			s.Equal(tt.out.Description, response.Description)
			s.Equal(tt.out.AmountEmployees, response.AmountEmployees)
			s.Equal(tt.out.Registered, response.Registered)
			s.Equal(tt.out.Type, response.Type)
		}
	})

	s.Run("unauthorized", func() {
		input := &schemas.CreateCompanyRequest{
			Name:            "test",
			Description:     "test",
			AmountEmployees: helpers.PointerValue(10),
			Registered:      helpers.PointerValue(true),
			Type:            enum.Corporation.String(),
		}

		jsonData, err := json.Marshal(input)
		s.Require().NoError(err)

		client := &http.Client{}
		req, err := http.NewRequest("POST", s.apiURL+"/company/create", bytes.NewBuffer(jsonData))
		s.Require().NoError(err)

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		s.Require().NoError(err)
		s.Equal(apierrors.ErrTokenNotFound.HTTPStatus, resp.StatusCode)
	})
}

func (s *integrationSuite) TestGetCompanyByID() {
	// login
	bodyLogin := `{"email":"` + s.email + `","password":"` + s.pasword + `"}`
	loginResp, err := http.Post(s.apiURL+"/login", "application/json", strings.NewReader(bodyLogin))
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, loginResp.StatusCode)

	var loginResponse schemas.LoginResponse
	err = json.NewDecoder(loginResp.Body).Decode(&loginResponse)
	s.Require().NoError(err)
	loginResp.Body.Close()

	// create a company
	company := &schemas.CreateCompanyRequest{
		Name:            "testGet",
		Description:     "test",
		AmountEmployees: helpers.PointerValue(10),
		Registered:      helpers.PointerValue(true),
		Type:            enum.Corporation.String(),
	}
	companyJSON, err := json.Marshal(company)
	s.Require().NoError(err)

	client := &http.Client{}
	req, err := http.NewRequest("POST", s.apiURL+"/company/create", bytes.NewBuffer(companyJSON))
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)

	resp, err := client.Do(req)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var createdCompany models.CompanyModel
	err = json.NewDecoder(resp.Body).Decode(&createdCompany)
	s.Require().NoError(err)
	resp.Body.Close()

	s.Run("ok", func() {
		input := []struct {
			in                 string
			out                models.CompanyModel
			expectedStatusCode int
		}{
			{
				in: createdCompany.ID.String(),
				out: models.CompanyModel{
					ID:              createdCompany.ID,
					Name:            company.Name,
					Description:     company.Description,
					AmountEmployees: *company.AmountEmployees,
					Registered:      *company.Registered,
					Type:            company.Type,
				},
				expectedStatusCode: http.StatusOK,
			},
		}

		for _, tt := range input {
			uri := fmt.Sprintf("%s/company/%s", s.apiURL, tt.in)
			resp, err := http.Get(uri)
			s.Require().NoError(err)
			s.Equal(tt.expectedStatusCode, resp.StatusCode)

			var response models.CompanyModel
			err = json.NewDecoder(resp.Body).Decode(&response)
			s.Require().NoError(err)

			s.Equal(tt.out.ID, response.ID)
			s.Equal(tt.out.Name, response.Name)
			s.Equal(tt.out.Description, response.Description)
			s.Equal(tt.out.AmountEmployees, response.AmountEmployees)
			s.Equal(tt.out.Registered, response.Registered)
			s.Equal(tt.out.Type, response.Type)
		}

	})
}

func (s *integrationSuite) TestUpdateCompany() {
	// login
	bodyLogin := `{"email":"` + s.email + `","password":"` + s.pasword + `"}`
	loginResp, err := http.Post(s.apiURL+"/login", "application/json", strings.NewReader(bodyLogin))
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, loginResp.StatusCode)

	var loginResponse schemas.LoginResponse
	err = json.NewDecoder(loginResp.Body).Decode(&loginResponse)
	s.Require().NoError(err)
	loginResp.Body.Close()

	// create a company
	company := &schemas.CreateCompanyRequest{
		Name:            "testUpdate",
		Description:     "test",
		AmountEmployees: helpers.PointerValue(10),
		Registered:      helpers.PointerValue(true),
		Type:            enum.Corporation.String(),
	}
	companyJSON, err := json.Marshal(company)
	s.Require().NoError(err)

	client := &http.Client{}
	req, err := http.NewRequest("POST", s.apiURL+"/company/create", bytes.NewBuffer(companyJSON))
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)

	resp, err := client.Do(req)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var createdCompany models.CompanyModel
	err = json.NewDecoder(resp.Body).Decode(&createdCompany)
	s.Require().NoError(err)
	resp.Body.Close()

	s.Run("ok", func() {
		input := []struct {
			id                 string
			in                 *schemas.UpdateCompanyRequest
			out                models.CompanyModel
			expectedStatusCode int
		}{
			{
				id: createdCompany.ID.String(),
				in: &schemas.UpdateCompanyRequest{
					Name:            "testUpdate2",
					Description:     "test2",
					AmountEmployees: helpers.PointerValue(20),
					Registered:      helpers.PointerValue(false),
					Type:            enum.Cooperative.String(),
				},
				expectedStatusCode: http.StatusOK,
			},
		}

		for _, tt := range input {
			uri := fmt.Sprintf("%s/company/%s", s.apiURL, tt.id)
			fmt.Println(uri)
			jsonData, err := json.Marshal(tt.in)
			s.Require().NoError(err)

			client := &http.Client{}
			req, err := http.NewRequest("PUT", uri, bytes.NewBuffer(jsonData))
			s.Require().NoError(err)

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)

			resp, err := client.Do(req)
			s.Require().NoError(err)
			s.Equal(tt.expectedStatusCode, resp.StatusCode)
		}
	})
}

func (s *integrationSuite) TestDeleteCompany() {
	// login
	bodyLogin := `{"email":"` + s.email + `","password":"` + s.pasword + `"}`
	loginResp, err := http.Post(s.apiURL+"/login", "application/json", strings.NewReader(bodyLogin))
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, loginResp.StatusCode)

	var loginResponse schemas.LoginResponse
	err = json.NewDecoder(loginResp.Body).Decode(&loginResponse)
	s.Require().NoError(err)
	loginResp.Body.Close()

	// create a company
	company := &schemas.CreateCompanyRequest{
		Name:            "testDelete",
		Description:     "test",
		AmountEmployees: helpers.PointerValue(10),
		Registered:      helpers.PointerValue(true),
		Type:            enum.Corporation.String(),
	}
	companyJSON, err := json.Marshal(company)
	s.Require().NoError(err)

	client := &http.Client{}
	req, err := http.NewRequest("POST", s.apiURL+"/company/create", bytes.NewBuffer(companyJSON))
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)

	resp, err := client.Do(req)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var createdCompany models.CompanyModel
	err = json.NewDecoder(resp.Body).Decode(&createdCompany)
	s.Require().NoError(err)
	resp.Body.Close()

	s.Run("ok", func() {
		uri := fmt.Sprintf("%s/company/%s", s.apiURL, createdCompany.ID.String())
		req, err := http.NewRequest("DELETE", uri, nil)
		s.Require().NoError(err)

		req.Header.Set("Authorization", "Bearer "+loginResponse.AccessToken)

		resp, err := client.Do(req)
		s.Require().NoError(err)
		s.Equal(http.StatusOK, resp.StatusCode)
	})
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(integrationSuite))
}
