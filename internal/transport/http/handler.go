package http

import (
	"fmt"
	"net/http"
	"time"
	apierrors "xm_test/internal/api_errors"
	"xm_test/internal/db"
	"xm_test/internal/enum"
	"xm_test/internal/events"
	"xm_test/internal/service"
	"xm_test/internal/service/inputs"
	"xm_test/internal/transport/http/binding"
	"xm_test/internal/transport/http/schemas"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type handler struct {
	logger *zap.SugaredLogger
	db     db.DatabaseAdapter

	// services
	as service.AuthService
	cs service.CompanyService

	// event dispatcher
	evtDispatcher events.Dispatcher
}

// newHandler creates a new handler.
func newHandler(logger *zap.SugaredLogger, db db.DatabaseAdapter) *handler {
	// initiate services
	as := service.NewAuthService(logger, db)
	cs := service.NewCompanyService(logger, db)

	dispatcher := events.NewEventsDispatcher(logger, db)
	return &handler{logger: logger, db: db, as: as, cs: cs, evtDispatcher: dispatcher}
}

// Register registers a new user
func (h *handler) register(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("registering user endpoint called")

	h.logger.Debugf("decoding request body")
	var body schemas.RegisterRequest
	if err := binding.DecodeJSONBody(r, &body); err != nil {
		e := apierrors.ErrInvalidBody
		e.Message = fmt.Sprintf("failed to decode request body: %v", err)
		h.wrapError(w, r, e)
		return
	}
	h.logger.Debugf("request body decoded")

	h.logger.Debugf("creating account for user with email '%s'", body.Email)
	if err := h.as.Register(body.Email, body.Password); err != nil {
		h.wrapError(w, r, err)
		return
	}
	h.logger.Infof("user with email '%s' registered", body.Email)
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, schemas.OkResponse{Message: "user registered"})
}

// Login logs in a user
func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("login endpoint called")

	h.logger.Debugf("decoding request body")
	var body schemas.LoginRequest
	if err := binding.DecodeJSONBody(r, &body); err != nil {
		e := apierrors.ErrInvalidBody
		e.Message = fmt.Sprintf("failed to decode request body: %v", err)
		h.wrapError(w, r, e)
		return
	}
	h.logger.Debugf("request body decoded")

	h.logger.Debugf("logging in user with email '%s'", body.Email)
	token, err := h.as.Login(body.Email, body.Password)
	if err != nil {
		h.wrapError(w, r, err)
		return
	}
	h.logger.Infof("user with email '%s' logged in", body.Email)
	render.JSON(w, r, schemas.LoginResponse{AccessToken: *token})
}

// CreateCompany creates a new company
func (h *handler) createCompany(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("create company endpoint called")

	h.logger.Debugf("decoding request body")
	var body schemas.CreateCompanyRequest
	if err := binding.DecodeJSONBody(r, &body); err != nil {
		e := apierrors.ErrInvalidBody
		e.Message = fmt.Sprintf("failed to decode request body: %v", err)
		h.wrapError(w, r, e)
		return
	}
	h.logger.Debugf("request body decoded")

	h.logger.Debugf("creating company with name '%s'", body.Name)
	input := &inputs.CreateCompanyInput{
		Name:            body.Name,
		Description:     body.Description,
		AmountEmployees: body.AmountEmployees,
		Registered:      body.Registered,
		Type:            body.Type,
	}
	companyModel, err := h.cs.CreateCompany(input)
	if err != nil {
		h.wrapError(w, r, err)
		return
	}

	// create event
	go h.createEvent(&events.Event{
		Type:      enum.EventCreateCompany.String(),
		Timestamp: time.Now(),
		ID:        uuid.New(),
		EntityID:  companyModel.ID,
	})

	h.logger.Infof("company with name '%s' created", body.Name)
	render.JSON(w, r, companyModel)
}

// GetCompany retrieves a company by its ID
func (h *handler) getCompany(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("get company endpoint called")

	h.logger.Debugf("decoding company id from the request")
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		e := apierrors.ErrCompanyIDRequired
		h.wrapError(w, r, e)
		return
	}

	// check if the company id is a valid uuid
	if err := uuid.Validate(companyID); err != nil {
		e := apierrors.ErrInvalidUUID
		h.wrapError(w, r, e)
		return
	}
	h.logger.Debugf("company id decoded: %s", companyID)

	h.logger.Debugf("retrieving company with id '%s'", companyID)
	company, err := h.cs.GetCompanyByID(companyID)
	if err != nil {
		h.wrapError(w, r, err)
		return
	}
	h.logger.Infof("company with id '%s' retrieved", companyID)
	render.Status(r, http.StatusOK)
	render.JSON(w, r, company)
}

// UpdateCompany updates a company
func (h *handler) updateCompany(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("update company endpoint called")

	h.logger.Debugf("decoding company id from the request")
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		e := apierrors.ErrCompanyIDRequired
		h.wrapError(w, r, e)
		return
	}

	// check if the company id is a valid uuid
	if err := uuid.Validate(companyID); err != nil {
		e := apierrors.ErrInvalidUUID
		h.wrapError(w, r, e)
		return
	}
	h.logger.Debugf("company id decoded: %s", companyID)

	h.logger.Debugf("decoding request body")
	var body schemas.UpdateCompanyRequest
	if err := binding.DecodeJSONBody(r, &body); err != nil {
		e := apierrors.ErrInvalidBody
		e.Message = fmt.Sprintf("failed to decode request body: %v", err)
		h.wrapError(w, r, e)
		return
	}
	h.logger.Debugf("request body decoded")

	h.logger.Debugf("updating company with id '%s'", companyID)
	input := &inputs.UpdateCompany{
		Name:            body.Name,
		Description:     body.Description,
		AmountEmployees: body.AmountEmployees,
		Registered:      body.Registered,
		Type:            body.Type,
	}
	if err := h.cs.UpdateCompany(companyID, input); err != nil {
		h.wrapError(w, r, err)
		return
	}

	go h.createEvent(&events.Event{
		Type:      enum.EventUpdateCompany.String(),
		Timestamp: time.Now(),
		ID:        uuid.New(),
		EntityID:  uuid.MustParse(companyID),
	})

	h.logger.Infof("company with id '%s' updated", companyID)
	render.JSON(w, r, schemas.OkResponse{Message: "company updated"})
}

// DeleteCompany deletes a company
func (h *handler) deleteCompany(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("delete company endpoint called")

	h.logger.Debugf("decoding company id from the request")
	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		e := apierrors.ErrCompanyIDRequired
		h.wrapError(w, r, e)
		return
	}

	// check if the company id is a valid uuid
	if err := uuid.Validate(companyID); err != nil {
		e := apierrors.ErrInvalidUUID
		h.wrapError(w, r, e)
		return
	}
	h.logger.Debugf("company id decoded: %s", companyID)

	h.logger.Debugf("deleting company with id '%s'", companyID)
	if err := h.cs.DeleteCompany(companyID); err != nil {
		h.wrapError(w, r, err)
		return
	}

	go h.createEvent(&events.Event{
		Type:      enum.EventDeleteCompany.String(),
		Timestamp: time.Now(),
		ID:        uuid.New(),
		EntityID:  uuid.MustParse(companyID),
	})

	h.logger.Infof("company with id '%s' deleted", companyID)
	render.JSON(w, r, schemas.OkResponse{Message: "company deleted"})
}

func (h *handler) createEvent(evt *events.Event) {
	h.logger.Debugf("creating event for company created")
	if err := h.evtDispatcher.Dispatch(evt); err != nil {
		h.logger.Errorf("failed to dispatch event: %v", err)
		return
	}
	h.logger.Debugf("event for company created")
}

// wrapError logs the error and writes it to the response.
func (h *handler) wrapError(w http.ResponseWriter, r *http.Request, err error) {
	apiError, ok := err.(*apierrors.APIError)
	if !ok {
		unknownError := apierrors.ErrInternalServer
		unknownError.Message = err.Error()
		h.logger.Error(unknownError)
		render.JSON(w, r, unknownError)
		return
	}

	h.logger.Error(apiError.Message)
	render.Status(r, apiError.HTTPStatus)
	render.JSON(w, r, apiError)
}
