package conf

import (
	"fmt"
	"xm_test/internal/enum"

	"github.com/spf13/viper"
)

// setupConfig is a function that sets up the configuration for the application by reading from environment variables
// and validating the configuration. The configuration is stored in the conf package.
func SetupConfig() error {
	setDefaults()

	// Read from environment variables
	viper.AutomaticEnv()

	cfg := NewConfig()
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("bootstrap: config: failed to unmarshal configuration: %v", err)
	}

	// set database configuration
	if err := setDatabaseConfig(cfg); err != nil {
		return err
	}

	return cfg.Validate()
}

func setDatabaseConfig(cfg *Config) error {
	switch cfg.DatabaseType {
	case enum.Postgres:
		var postgres Postgres
		if err := viper.Unmarshal(&postgres); err != nil {
			return fmt.Errorf("bootstrap: config: failed to unmarshal postgres configuration: %v", err)
		}
		cfg.Postgres = postgres
	default:
		return fmt.Errorf("bootstrap: config: unsupported database type: %s", cfg.DatabaseType)
	}

	return nil

}

// setDefaults is a function that sets the default values for the API configuration.
func setDefaults() {
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("HEALTH_PORT", "8081")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("DATABASE_TYPE", "postgres")
	viper.SetDefault("JWT_SECRET", "secret")

	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_PORT", "5432")
	viper.SetDefault("POSTGRES_USER", "postgres")
	viper.SetDefault("POSTGRES_PASSWORD", "password")
	viper.SetDefault("POSTGRES_DATABASE", "xm")
	viper.SetDefault("POSTGRES_INIT_SCRIPT", "_db_schema/postgres/schema.sql")
}
