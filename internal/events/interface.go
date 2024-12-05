package events

import (
	"time"
	"xm_test/internal/db"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

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

// Event is the interface that defines the methods that the dispatcher must implement
type Dispatcher interface {
	// Dispatch dispatches an event to kafka, rabbitmq, or any other event bus, and it stores the event in the database
	Dispatch(event *Event) error
}

// NewEventsDispatcher returns a new events dispatcher instance
func NewEventsDispatcher(logger *zap.SugaredLogger, db db.DatabaseAdapter) Dispatcher {
	return newEventHandler(logger, db)
}
