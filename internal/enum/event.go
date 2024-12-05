package enum

// EventType represents the type of an event
type EventType string

const (
	EventCreateCompany EventType = "create_company"
	EventUpdateCompany EventType = "update_company"
	EventDeleteCompany EventType = "delete_company"
)

// String returns the string representation of the event type
func (e EventType) String() string {
	return string(e)
}

// IsValid checks if the event type is valid
func (e EventType) IsValid() bool {
	switch e {
	case EventCreateCompany, EventUpdateCompany, EventDeleteCompany:
		return true
	}
	return false
}
