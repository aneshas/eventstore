package eventstore

import "time"

// EventToStore represents an event that is to be stored in the event store
type EventToStore struct {
	Event any

	// Optional
	ID                 string
	CausationEventID   string
	CorrelationEventID string
	Meta               map[string]string
	OccurredOn         time.Time
}

// StoredEvent holds stored event data and meta data
type StoredEvent struct {
	Event any
	Meta  map[string]string

	ID                 string
	Sequence           uint64
	Type               string
	CausationEventID   *string
	CorrelationEventID *string
	StreamID           string
	StreamVersion      int
	OccurredOn         time.Time
}
