# Test API

## Overview

This repository aims to build a microservice for handling companies. It must provide the following operations:

- Create a company
- Patch a company
- Delete a company
- Get a company

## Design

This section describes how the project has been structured and the line of thought that I have followed to accomplish the definition of the API.

Below, it can be seen the project's structure

```md
.
├── cmd
│   ├── bootstrap
│   │   ├── bootstrap.go    
│   │   └── log.go
│   └── main.go 
├── internal
│   ├── api_errors
│   ├── crypto
│   ├── conf
│   ├── db
│   ├── enum
│   ├── events
│   ├── helpers
│   ├── mocks
│   ├── projectpath
│   ├── service
│   │   ├── auth
│   │   ├── company
│   │   ├── inputs
│   │   └── interface.go 
│   ├── token
│   └── transport
│       ├── http
│       └── interface.go              
├── tests
├── .air.toml
├── .dockerignore
├── .env
├── .gitignore
├── docker-compose.yaml
├── Dockerfile
├── go.mod
├── go.sum
├── README.md
└── Taskfile.yaml
```

The `cmd` folder contains the `main.go` file, which starts the API. If you look at this file, you'll see that its sole purpose is to call the `Run()` method defined in the bootstrap folder. This function is responsible for setting up the API. Specifically, it handles:

- Configuring the application settings by reading the configuration from environmental variables.
- Initializing the logger used for API logging.
- Starting the database. In this case, a Postgres database has been used to store the companies' data.
- Launching the HTTP server.

To configure the application, it was decided that the most optimal approach would be to use environment variables. This ensures that users can easily modify the tool's settings. Additionally, deploying the API in Docker or Kubernetes simplifies configuration management, as environment variables are easy to define in these platforms. The `.env` file contains the environment variables used by the application.

```bash
PORT=3000 # Define the port in which the API will run
HEALTH_PORT=3001 # Define the port in which the health check will run
LOG_LEVEL=debug # Define the log level of the API. It can be debug or info 
DATABASE_TYPE=postgres # Define the database type. Only postgres is supported

# secret
JWT_SECRET="this a secret key used to validate the jwt"

# postgres options
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_PASSWORD=xm
POSTGRES_USER=xm
POSTGRES_DB=xm
POSTGRES_INIT_SCRIPT=_db_schema/postgres/schema.sql
```

As you can see in the `.env` file, two ports are specified: one for the API to handle requests and another for the health check. The decision to use a separate port for the health check allows monitoring systems to independently verify the service's health without accessing the main API endpoints. This approach ensures the application remains operational while minimizing the risk of overloading the primary API or exposing sensitive information.

Additionally, the environmental variables `JWT_SECRET` and `DATABASE_TYPE` sets up the secret used to sign the JWT, and the technology used in the database layer respectively.

The folder `internal` contains all the logic of the API. Since no packages are going to be externalized, it makes sense to defined all the packages here.1

`api_errors` defines standard errors that can be returned by the API. All errors are represented by the following structured. Moreover, some common errors have already been defined.

```go
// APIError represents an error that is returned to the client.
type APIError struct {
	Code       string `json:"code"`    // error code that can be used to identify the error
	Message    string `json:"message"` // detailed description of the error
	HTTPStatus int    `json:"-"`       // http status code. It is not included in the response body
}
```

```go
// Common error definitions
var (
	// ErrInternalServer is returned when an internal server error occurs.
	ErrInternalServer = NewAPIError("INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
	// ErrInvalidBody is returned when the request body is invalid.
	ErrInvalidBody = NewAPIError("INVALID_BODY", "invalid request body", http.StatusBadRequest)
	// ErrUserNotFound is returned when a user is not found.
	ErrUserNotFound = NewAPIError("USER_NOT_FOUND", "user not found", http.StatusBadRequest)
	// ErrUserAlreadyExists is returned when a user already exists.
	ErrUserAlreadyExists = NewAPIError("USER_ALREADY_EXISTS", "user already exists", http.StatusBadRequest)
	// ErrInvalidCredentials is returned when the credentials are invalid.
	ErrInvalidCredentials = NewAPIError("INVALID_CREDENTIALS", "invalid credentials", http.StatusUnauthorized)
	// ErrCompanyNotFound is returned when a company is not found.
	ErrCompanyNotFound = NewAPIError("COMPANY_NOT_FOUND", "company not found", http.StatusBadRequest)
	// ErrInvalidUUID is returned when the UUID is invalid.
	ErrInvalidUUID = NewAPIError("INVALID_UUID", "invalid UUID", http.StatusBadRequest)
	// ErrTokenNotFound is returned when a token is not found.
	ErrTokenNotFound = NewAPIError("TOKEN_NOT_FOUND", "token not found", http.StatusBadRequest)
	// ErrInvalidToken is returned when the token is invalid.
	ErrInvalidToken = NewAPIError("INVALID_TOKEN", "invalid token", http.StatusUnauthorized)
	// ErrUnauthorized is returned when the user is not authorized.
	ErrUnauthorized = NewAPIError("UNAUTHORIZED", "unauthorized", http.StatusUnauthorized)
	// ErrTokenExpired is returned when the token is expired.
	ErrTokenExpired = NewAPIError("TOKEN_EXPIRED", "token expired", http.StatusUnauthorized)
	// ErrCompanyIDRequired is returned when the company ID is required.
	ErrCompanyIDRequired = NewAPIError("COMPANY_ID_REQUIRED", "company ID is required", http.StatusBadRequest)
	// ErrCreatingEvent is returned when an error occurs while creating an event.
	ErrCreatingEvent = NewAPIError("CREATING_EVENT", "error creating event", http.StatusInternalServerError)
)
```

in the folder `conf` you can see how the API reads its configuration, specifically this package reads the environmental variables and loads them in global variable called `GlobalConfig`. The package [viper](https://github.com/spf13/viper) has been used to load the configuration in the API.

```go
var GlobalConfig *Config

// Postgres holds the configuration values for a Postgres database
type Postgres struct {
	Host       string `mapstructure:"POSTGRES_HOST" validate:"required"`
	Port       string `mapstructure:"POSTGRES_PORT" validate:"required"`
	User       string `mapstructure:"POSTGRES_USER" validate:"required"`
	Password   string `mapstructure:"POSTGRES_PASSWORD" validate:"required"`
	Database   string `mapstructure:"POSTGRES_DATABASE" validate:"required"`
	InitScript string `mapstructure:"POSTGRES_INIT_SCRIPT" validate:"required"`
}

// Config holds the configuration values for the API
type Config struct {
	Port       string        `mapstructure:"PORT" validate:"required"`        // Port in which the API will listen
	HealthPort string        `mapstructure:"HEALTH_PORT" validate:"required"` // Health port in which the API will listen
	LogLevel   enum.LogLevel `mapstructure:"LOG_LEVEL" validate:"required"`   // Log level for the API: debug, info
	JwtSecret  string        `mapstructure:"JWT_SECRET" validate:"required"`  // JWT secret key

	DatabaseType enum.DatabaseType `mapstructure:"DATABASE_TYPE" validate:"required"` // Database type. Default: postgres
	Postgres     Postgres          // Database configuration
}
```

As seen in the previous Go script, the `Config` struct contains a field named `DatabaseType`. This field specifies which underlying technology the API should use for the persistence layer. Currently, Postgres is being used. However, if another technology needs to be implemented in the future, the `db` package provides an interface that facilitates this process, as will be shown later. In such cases, you would only need to update the environment variable and create a new struct with the necessary database parameters.

The package `crypto` contains a simple function to hash the users passwords with the MD5 algorithm, so they can be stored securely in the database. The MD5 algorithm has been chosen for its simplicity. In a production environment you should implement a more secured algorithm.

The folder `db` contains the package used to interact with the persistence layer. In this case, keeping in mind that APIs might required changes in the future, an interface has been defined to interact with the database. This interface indicates all the operations performed against the persistence layer.

```go
// DatabaseAdapter is the interface that wraps the basic methods to interact with the database layer.
//
// By using this interface, we can easily swap out the underlying database implementation.
type DatabaseAdapter interface {
	// Connection operations
	Connect(ctx context.Context, opts ...func(*options.DatabaseOptions)) error
	Close(ctx context.Context) error

	// auth table operations
	CreateUser(ctx context.Context, user *models.UserModel) error
	GetUserByEmail(ctx context.Context, email string) (*models.UserModel, error)

	// company table operations
	CreateCompany(ctx context.Context, company *models.CompanyModel) error
	GetCompanyByID(ctx context.Context, id string) (*models.CompanyModel, error)
	UpdateCompany(ctx context.Context, id string, updateCompany *models.CompanyModel) error
	DeleteCompany(ctx context.Context, id string) error

	// events table operations
	CreateEvent(ctx context.Context, event *models.EventModel) error
}
```

You can implement any database technology as long as it implements the previous interface. Currently, only Postgres is supported.

```go
// NewDatabaseAdapter returns a new DatabaseAdapter instance.
func NewDatabaseAdapter(logger *zap.SugaredLogger, opts ...func(*options.DatabaseOptions)) DatabaseAdapter {
	switch conf.GlobalConfig.DatabaseType {
	case enum.Postgres:
		db := postgres.NewPostgresAdapter(logger)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.Connect(ctx, opts...); err != nil {
			logger.Fatalf("failed to connect to database: %s", err)
			return nil
		}
		return db
	default:
		logger.Fatalf("database type '%s' not supported", conf.GlobalConfig.DatabaseType)
		return nil
	}
}
```

Moreover, this package contains the data models that will be stored in the database. Specifically, three structs have been defined:

1. `CompanyModel`: Represents a company in the database.
2. `UserModel`: Represents a user in the database. One of the problem requirements was that some endpoints were restricted to logged in users. Thus, it is necessary to store user's accounts in the database, so they can log in in the tool.
3. `EventModel`: Represents an event in the database.

```go
// CompanyModel represents the company model
type CompanyModel struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	AmountEmployees int       `json:"amount_employees" db:"amount_employees"`
	Registered      bool      `json:"registered" db:"registered"`
	Type            string    `json:"type" db:"type"`
}

// UserModel represents the user model
type UserModel struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Email       string    `json:"email" db:"email"`
	EncPassword string    `json:"enc_password" db:"enc_password"`
}

// EventModel represents the event model
type EventModel struct {
	Type      string             `json:"type" db:"type"`
	Timestamp pgtype.Timestamptz `json:"timestamp" db:"timestamp"`
	ID        uuid.UUID          `json:"id" db:"id"`
	EntityID  uuid.UUID          `json:"entity_id" db:"entity_id"`
}
```

Postgres has been chosen as the underlying database technology. The package [pgx](https://pkg.go.dev/github.com/jackc/pgx/v5) is used to interact with the database. For example, the next portion of code shows how companies are stored in the database.

```go
// CreateCompany is a method that creates a new company in the database.
func (p *postgresDB) CreateCompany(ctx context.Context, company *models.CompanyModel) error {
	p.logger.Debugf("creating company: %s", company.Name)
	args := pgx.NamedArgs{
		"id":               company.ID.String(),
		"name":             company.Name,
		"description":      company.Description,
		"amount_employees": company.AmountEmployees,
		"registered":       company.Registered,
		"type":             company.Type,
	}
	cmd := "INSERT INTO company (id, name, description, amount_employees, registered, type) VALUES (@id, @name, @description, @amount_employees, @registered, @type)"
	p.logger.Debugf("cmd: %s", cmd)

	if _, err := p.client.Exec(ctx, cmd, args); err != nil {
		apiError := apierrors.ErrInternalServer
		apiError.Message = fmt.Sprintf("failed to create company: %s", err)
		return apiError
	}
	p.logger.Debugf("created company: %s", company.Name)
	return nil
}
```
The package `enum` contains enum definitions used by the API. Specifically, for definitions:

- `CompanyType`: indicates the possible types that a company can have.
  - Corporations
  - NonProfit
  - Cooperative
  - Sole Proprietorship
- `DatabaseType`: indicates the databases supported by the API. At the moment, it only supports `postgres`
- `EventType`: represents the possible events types. For more information check the description of the `events` package.
  - create_company
  - update_company
  - delete_company
- `LogLevel`: Indicates the log level of the api. Only `debug` and `info` levels are supported.

The package `events` contains an event dispatcher. This module is triggered when a successful entity modification has been made in the database. It process the event; sends it to a queue, and it stores it in the database. If you take a look at the package, you will see that this element is represented by the following interface.

```go
// Event is the interface that defines the methods that the dispatcher must implement
type Dispatcher interface {
	// Dispatch dispatches an event to kafka, rabbitmq, or any other event bus, and it stores the event in the database
	Dispatch(event *Event) error
}
```

The struct `eventHandler` implements the previous interface. As it is shown in the next piece of code, the event handler sends the event to a queue (**This has not been implemented due to the lack of time**), and it stores the event in the database

```go
// Dispatch dispatches an event to kafka, rabbitmq, or any other event bus, and it stores the event in the database
func (e *eventHandler) Dispatch(event *Event) error {
	// dispatch event to event bus
	e.logger.Infof("dispatching event '%s' to event bus with ID '%s' at '%s'", event.Type, event.ID, event.Timestamp.Format(time.RFC3339))
	// TODO: here you would dispatch the event to an event bus like kafka, rabbitmq, etc.
	e.logger.Infof("event with ID '%s' dispatched", event.ID)

	// store event in database
	e.logger.Debugf("storing event '%s' in database", event.ID)
	eventModel := &models.EventModel{
		Type:      event.Type,
		Timestamp: pgtype.Timestamptz{Time: event.Timestamp, Valid: true},
		ID:        event.ID,
		EntityID:  event.EntityID,
	}
	if err := e.db.CreateEvent(context.Background(), eventModel); err != nil {
		e := apierrors.ErrCreatingEvent
		e.Message = fmt.Sprintf("failed to store event in database: %s", err)
		return e
	}
	e.logger.Debugf("event '%s' stored in database", event.ID)
	return nil
}
```

Events are represented using the following struct.

```go
// Events are used to track changes in the system.
// They are triggered when a change is made to the database:
//
// - Create: When a company is created in the database
//
// - Update: When a company is updated in the database
//
// - Delete: When a company is deleted from the database
type Event struct {
	Type      string    `json:"type" db:"type"`           // create_company, update_company, delete_company
	Timestamp time.Time `json:"timestamp" db:"timestamp"` // The time the event was created
	ID        uuid.UUID `json:"id" db:"id"`               // The unique identifier of the event
	EntityID  uuid.UUID `json:"entity_id" db:"entity_id"` // The unique identifier of the entity that the event is related to
}
```

The package `helpers` contains multiple support functions that are used in other packages. 

`mocks` contains the code to initialize a postgres database in a docker container using the library [testcontainers](https://golang.testcontainers.org). This containerized database is used for testing purposes.

The package `service` contains the business logic of the application. Here, two services have been defined to interact with the accounts and to interact with companies.

```go
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
```

These services are in charge of receiving information from the transport layer, processing it, and interacting with the database.

The `AuthService` service is necessary to implement an authentication system, as one of the project's requirements is that certain endpoints be accessible only to authenticated users. Ideally, the authentication layer should be integrated in other microservice or externalized in other tools such as Keycloak.

This API implements a simple authentication system using JWT and the HTTP Authorization header to include access tokens. When a user registers with the service, their email and hashed password are stored in the database. Upon login, the service verifies that the user exists in the database and that the provided credentials are valid. If the credentials are correct, the service returns a JWT access token, which must be included in the Authorization header for protected routes.

For routes requiring authentication, the API verifies the token's validity by checking its signature to ensure it was issued by the API and confirming that the token has not expired. By default, tokens expire after 10 minutes, but users can request new valid tokens as needed.

The package `token` includes methods to issue JWT tokens, validate them, and extract the token from the HTTP request.

Finally, the last package in the `internal` folder is `transport`. This package defines the application's transport layer. Like the database package, it provides an interface to represent this layer, enabling future extensions with additional transport options. At present, only HTTP has been implemented.

The framework [chi](https://github.com/go-chi/chi) has been used to implement the HTTP server. Additionally, to validate request's bodies the framework [validator](https://github.com/go-playground/validator]). This framework allows users to set multiple rules in the struct tags that can be used to validate the requests.

A middleware has been created to verify that users are authenticated when accessing protected endpoints. This middleware checks that the token is valid and not expired.

```go
func UserMustBeAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if the user is authenticated
		// if not, return an error

		// decode the token from the request header
		claims, err := token.DecodeTokenFromRequest(r)
		if err != nil {
			render.Status(r, err.(*apierrors.APIError).HTTPStatus)
			render.JSON(w, r, err)
			return
		}

		// check whether the token is not expired
		if time.Now().After(claims.ExpiresAt.Time) {
			e := apierrors.ErrTokenExpired
			render.Status(r, e.HTTPStatus)
			render.JSON(w, r, e)
			return
		}

		// add the claims to the request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, claimsKey, claims)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
```

In the folder tests, you can find an integration test of the API. This has been done by dockerizing the API and the postgres db using the library [testcontainers](https://golang.testcontainers.org/) and performing multiple queries to each endpoint of the API.

## Endpoints

The next sections describe the available endpoints in the API. When an endpoint is marked as **protected**, it means that the user must be authenticated to operate with it.

### Health

- `GET /health`: Healthcheck. **This endpoint is listening on the healtcheck port!!**

### Auth service

- `POST /register`: Registers a user in the database.
example body:

```json
{
    "email": "test@test.es",
    "password": "test"
}
```

- `POST /login`: logs in user in the database
Example body


```json
{
    "email": "test@test.es",
    "password": "test"
}
```

### Company service

- (**PROTECTED**) `POST /company/create`: Creates a new company

Example body

```json
{
    "name": "test67",
    "description": "this is a rando2m description",
    "amount_employees": 1,
    "registered": true,
    "type": "NonProfit"
}
```

- `GET /company/:company_id`: Gets company from its UUID.
- (**PROTECTED**) `PUT /company/:company_id`: Updates a company.

Example body:

```json
{
    "name": "test1",
    "description": "this is a rando2m description",
    "amount_employees": 20,
    "registered": true,
    "type": "NonProfit"
}
```

- (**PROTECTED**) `DELETE /company/:company_id`: Deletes a company.


## Installation and usage

The API can be launch using the tasks defined in the Taskfile.yaml, so the package [task](https://taskfile.dev/) must be installed in your computer. The next commands can be used to run tests and launch the API.

- Run API in development mode (hot reload) by using [Air](https://github.com/air-verse/air): `task dev`
- Run API: `task run`
- Run tests: `task test`
- Run integration tests: `task integration_test`
- Run API in docker compose: `task docker`
