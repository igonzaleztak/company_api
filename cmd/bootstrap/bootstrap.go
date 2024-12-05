package bootstrap

import (
	"log"
	"xm_test/internal/conf"
	"xm_test/internal/db"
	"xm_test/internal/helpers"
	"xm_test/internal/transport"
)

func Run() error {
	// Setup the configuration
	if err := conf.SetupConfig(); err != nil {
		return err
	}

	// Setup the logger
	logger, err := NewZapLogger()
	if err != nil {
		return err
	}

	logger.Info("Starting XM Test API")
	logger.Debugf("starting with config: %s", helpers.PrettyPrintStructResponse(conf.GlobalConfig))

	logger.Debugf("setting up database connection")
	db := db.NewDatabaseAdapter(logger)
	logger.Debugf("database connection established")

	// Setup the transport layer and start the server
	server := transport.NewTransporter(logger, db)

	go func() {
		if err := server.HealthCheck(); err != nil {
			logger.Error(err)
			log.Fatal(err)
		}
	}()

	if err := server.Serve(); err != nil {
		return err
	}

	return nil
}
