package events

import (
	"context"
	"fmt"
	"time"
	apierrors "xm_test/internal/api_errors"
	"xm_test/internal/db"
	"xm_test/internal/db/models"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type eventHandler struct {
	logger *zap.SugaredLogger
	db     db.DatabaseAdapter
}

// NewEventHandler returns a new event handler instance
func newEventHandler(logger *zap.SugaredLogger, db db.DatabaseAdapter) *eventHandler {
	return &eventHandler{
		logger: logger,
		db:     db,
	}
}

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
