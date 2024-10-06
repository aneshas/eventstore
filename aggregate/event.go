package aggregate

import "time"

// Event represents a domain event
type Event struct {
	ID         string
	E          any
	OccurredOn time.Time

	CausationEventID   *string
	CorrelationEventID *string
	Meta               map[string]string
}
