package db

import (
	"context"
	"time"
	"xm_test/internal/conf"
	"xm_test/internal/db/models"
	"xm_test/internal/db/options"
	"xm_test/internal/db/postgres"
	"xm_test/internal/enum"

	"go.uber.org/zap"
)

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
