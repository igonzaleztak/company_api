package http

import (
	"fmt"
	"net/http"
	"xm_test/internal/conf"
	"xm_test/internal/db"
	"xm_test/internal/events"
	"xm_test/internal/helpers"
	"xm_test/internal/transport/http/schemas"

	customMiddlewares "xm_test/internal/transport/http/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"go.uber.org/zap"
)

type httpTransport struct {
	logger *zap.SugaredLogger
	db     db.DatabaseAdapter

	evtDispatcher events.Dispatcher
}

// NewHttpTransport returns a new http transport instance
func NewHttpTransport(logger *zap.SugaredLogger, db db.DatabaseAdapter) *httpTransport {
	return &httpTransport{logger: logger, db: db}
}

// Serve is a function that sets up the http server. It listens on the port specified in the configuration.
func (h httpTransport) Serve() error {
	h.logger.Debugf("setting up http server")
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// setup the routes here
	handler := newHandler(h.logger, h.db)

	protectedRoutes := r.Group(func(r chi.Router) {
		r.Use(customMiddlewares.UserMustBeAuthenticated)
	})

	// auth routes
	r.Post("/register", handler.register)
	r.Post("/login", handler.login)

	// company routes
	r.Get("/company/{id}", handler.getCompany)
	protectedRoutes.Post("/company/create", handler.createCompany)
	protectedRoutes.Put("/company/{id}", handler.updateCompany)
	protectedRoutes.Delete("/company/{id}", handler.deleteCompany)

	port := fmt.Sprintf(":%s", conf.GlobalConfig.Port)
	h.logger.Infof("http server listening on port %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		return h.wrapError(err)
	}

	return nil
}

// HealthCheck is a function that sets up the health check endpoint. It listens on the health port specified in the configuration,
// which is different from the main port to allow for easier monitoring of the application.
func (h httpTransport) HealthCheck() error {
	h.logger.Debugf("setting up health check endpoint")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		h.logger.Debug("health check endpoint called")
		response := schemas.HealthResponse{Message: "OK"}
		h.logger.Debugf("got health check response: %s", helpers.PrettyPrintStructResponse(response))
		render.JSON(w, r, response)
	})

	port := fmt.Sprintf(":%s", conf.GlobalConfig.HealthPort)
	h.logger.Infof("health check server listening on port %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		return h.wrapError(err)
	}

	return nil
}

func (h httpTransport) Close() error {
	return nil
}

// wrapError is a helper function that logs the error and returns it
func (h httpTransport) wrapError(err error) error {
	h.logger.Error("error")
	return err
}
