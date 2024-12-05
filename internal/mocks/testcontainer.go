package mocks

import (
	"context"
	"path/filepath"
	"time"
	"xm_test/internal/projectpath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RunPostgresTestDatabaseContainer runs a test postgres database container.
func RunPostgresTestDatabaseContainer(user string, database string, password string, initScript string) (*postgres.PostgresContainer, *string, error) {
	pathToInitScript := filepath.Join(projectpath.Root, initScript)

	ctx := context.Background()
	ctr, err := postgres.Run(ctx,
		"postgres:17.0",
		postgres.WithDatabase(database),
		postgres.WithUsername(user),
		postgres.WithPassword(password),
		postgres.WithInitScripts(pathToInitScript),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, nil, err
	}

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, err
	}

	return ctr, &connStr, nil
}
