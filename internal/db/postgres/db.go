package postgres

import (
	"context"
	"errors"
	"fmt"
	apierrors "xm_test/internal/api_errors"
	"xm_test/internal/conf"
	"xm_test/internal/db/models"
	"xm_test/internal/db/options"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
)

// postgres is a struct that manages the postgres database connection.
type postgresDB struct {
	logger *zap.SugaredLogger

	client *pgxpool.Pool
	isConn bool
}

// NewPostgresAdapter returns a new postgres instance.
func NewPostgresAdapter(logger *zap.SugaredLogger) *postgresDB {
	return &postgresDB{logger: logger}
}

// Connect is a method that establishes a connection to the postgres database.
func (p *postgresDB) Connect(ctx context.Context, opts ...func(*options.DatabaseOptions)) error {
	options := &options.DatabaseOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// setting default connection string if not provided
	if options.ConnString == "" {
		username := conf.GlobalConfig.Postgres.User
		password := conf.GlobalConfig.Postgres.Password
		host := conf.GlobalConfig.Postgres.Host
		port := conf.GlobalConfig.Postgres.Port
		database := conf.GlobalConfig.Postgres.Database
		options.ConnString = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, host, port, database)
	}

	p.logger.Debugf("connecting to postgres database")

	pool, err := pgxpool.New(ctx, options.ConnString)
	if err != nil {
		apiError := apierrors.ErrInternalServer
		apiError.Message = fmt.Sprintf("failed to connect to postgres database: %s", err)
		return apiError
	}
	p.logger.Debugf("connected to postgres database: '%s'", pool.Config().ConnConfig.Database)

	cfg := pool.Config()
	cfg.AfterConnect = func(ctx context.Context, pgconn *pgx.Conn) error {
		pgxUUID.Register(pgconn.TypeMap())
		return nil
	}

	p.client = pool
	p.isConn = true
	return nil
}

// Close is a method that closes the connection to the postgres database.
func (p *postgresDB) Close(ctx context.Context) error {
	if !p.isConn {
		return nil
	}

	p.logger.Debugf("closing postgres connection")
	p.client.Close()
	p.logger.Debugf("closed postgres connection")
	p.isConn = false
	return nil
}

// CreateUser is a method that creates a new user in the database.
func (p *postgresDB) CreateUser(ctx context.Context, user *models.UserModel) error {
	p.logger.Debugf("creating user: %s", user.Email)
	args := pgx.NamedArgs{
		"id":           user.ID.String(),
		"email":        user.Email,
		"enc_password": user.EncPassword,
	}
	cmd := "INSERT INTO users (id, email, enc_password) VALUES (@id, @email, @enc_password)"
	p.logger.Debugf("cmd: %s", cmd)

	if _, err := p.client.Exec(ctx, cmd, args); err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Code == pgerrcode.UniqueViolation {
			apiError := apierrors.ErrUserAlreadyExists
			apiError.Message = fmt.Sprintf("user with email '%s' already exists", user.Email)
			return apiError
		}
		apiError := apierrors.ErrInternalServer
		apiError.Message = fmt.Sprintf("failed to create user: %s", err)
		return apiError
	}
	p.logger.Debugf("created user: %s", user.Email)
	return nil
}

// GetUserByEmail is a method that retrieves a user by email from the database.
func (p *postgresDB) GetUserByEmail(ctx context.Context, email string) (*models.UserModel, error) {
	p.logger.Debugf("retrieving user by email: %s", email)
	users := make([]models.UserModel, 0)
	cmd := "SELECT * FROM users WHERE email = $1 LIMIT 1"
	p.logger.Debugf("cmd: %s", cmd)

	if err := pgxscan.Select(ctx, p.client, &users, cmd, email); err != nil {
		apiError := apierrors.ErrInternalServer
		apiError.Message = fmt.Sprintf("failed to retrieve user by email: %s", err)
		return nil, apiError
	}
	if len(users) == 0 {
		apiError := apierrors.ErrUserNotFound
		apiError.Message = fmt.Sprintf("user with email '%s' not found", email)
		return nil, apiError
	}

	user := users[0]
	p.logger.Debugf("user found with id: %s", user.ID.String())
	p.logger.Debugf("retrieved user by email: %s", email)
	return &user, nil
}

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

// GetCompanyByID is a method that retrieves a company by id from the database.
func (p *postgresDB) GetCompanyByID(ctx context.Context, id string) (*models.CompanyModel, error) {
	p.logger.Debugf("retrieving company by id: %s", id)
	companies := make([]models.CompanyModel, 0)
	cmd := "SELECT * FROM company WHERE id = $1 LIMIT 1"
	p.logger.Debugf("cmd: %s", cmd)

	if err := pgxscan.Select(ctx, p.client, &companies, cmd, id); err != nil {
		apiError := apierrors.ErrInternalServer
		apiError.Message = fmt.Sprintf("failed to retrieve company by id: %s", err)
		return nil, apiError
	}
	if len(companies) == 0 {
		apiError := apierrors.ErrCompanyNotFound
		apiError.Message = fmt.Sprintf("company with id '%s' not found", id)
		return nil, apiError
	}

	company := companies[0]
	p.logger.Debugf("company found with id: %s", company.ID.String())
	p.logger.Debugf("retrieved company by id: %s", id)
	return &company, nil
}

// UpdateCompany is a method that updates a company in the database.
func (p *postgresDB) UpdateCompany(ctx context.Context, id string, updateCompany *models.CompanyModel) error {
	p.logger.Debugf("updating company by id: %s", id)
	args := pgx.NamedArgs{
		"id":               id,
		"name":             updateCompany.Name,
		"description":      updateCompany.Description,
		"amount_employees": updateCompany.AmountEmployees,
		"registered":       updateCompany.Registered,
		"type":             updateCompany.Type,
	}
	cmd := "UPDATE company SET name = @name, description = @description, amount_employees = @amount_employees, registered = @registered, type = @type WHERE id = @id"
	p.logger.Debugf("cmd: %s", cmd)

	if _, err := p.client.Exec(ctx, cmd, args); err != nil {
		apiError := apierrors.ErrInternalServer
		apiError.Message = fmt.Sprintf("failed to update company by id: %s", err)
		return apiError
	}
	p.logger.Debugf("updated company by id: %s", id)
	return nil
}

// DeleteCompany is a method that deletes a company by id from the database.
func (p *postgresDB) DeleteCompany(ctx context.Context, id string) error {
	p.logger.Debugf("deleting company by id: %s", id)
	cmd := "DELETE FROM company WHERE id = $1"
	p.logger.Debugf("cmd: %s", cmd)

	if _, err := p.client.Exec(ctx, cmd, id); err != nil {
		apiError := apierrors.ErrInternalServer
		apiError.Message = fmt.Sprintf("failed to delete company by id: %s", err)
		return apiError
	}
	p.logger.Debugf("deleted company by id: %s", id)
	return nil
}

// CreateEvent is a method that creates a new event in the database.
func (p *postgresDB) CreateEvent(ctx context.Context, event *models.EventModel) error {
	p.logger.Debugf("creating event: %s", event.Type)
	args := pgx.NamedArgs{
		"id":        event.ID.String(),
		"type":      event.Type,
		"timestamp": event.Timestamp,
		"entity_id": event.EntityID.String(),
	}
	cmd := "INSERT INTO events (id, type, timestamp, entity_id) VALUES (@id, @type, @timestamp, @entity_id)"
	p.logger.Debugf("cmd: %s", cmd)

	if _, err := p.client.Exec(ctx, cmd, args); err != nil {
		apiError := apierrors.ErrInternalServer
		apiError.Message = fmt.Sprintf("failed to create event: %s", err)
		return apiError
	}
	p.logger.Debugf("created event: %s", event.Type)
	return nil
}
