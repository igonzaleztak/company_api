package conf

import (
	"fmt"
	"xm_test/internal/enum"

	"github.com/go-playground/validator/v10"
)

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

// NewConfig returns a new Config instance
func NewConfig() *Config {
	GlobalConfig = new(Config)
	return GlobalConfig
}

// Validate validates that all mandatory fields are correctly set
func (c *Config) Validate() error {
	// check log level enum
	if !c.LogLevel.IsValid() {
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}

	// check database type enum
	if !c.DatabaseType.IsValid() {
		return fmt.Errorf("invalid database type: %s", c.DatabaseType)
	}

	v := validator.New()
	return v.Struct(c)
}
