package aggregate

import (
	"context"
	"github.com/aneshas/eventstore"
)

// NewStore constructs new event sourced aggregate store
func NewStore[T Rooter](eventStore EventStore) *Store[T] {
	return &Store[T]{
		eventStore: eventStore,
	}
}

// EventStore represents event store
type EventStore interface {
	AppendStream(ctx context.Context, id string, version int, events []eventstore.EventToStore) error
	ReadStream(ctx context.Context, id string) ([]eventstore.StoredEvent, error)
}

// Store represents event sourced aggregate store
type Store[T Rooter] struct {
	eventStore EventStore
}

// Save saves aggregate events to the event store
func (s *Store[T]) Save(ctx context.Context, aggregate T) error {
	var events []eventstore.EventToStore

	for _, evt := range aggregate.Events() {
		events = append(events, eventstore.EventToStore{
			Event:      evt,
			ID:         evt.ID,
			OccurredOn: evt.OccurredOn,

			// Optional - set through context
			CausationEventID:   "",
			CorrelationEventID: "",
			Meta:               nil,
		})
	}

	return s.eventStore.AppendStream(
		ctx,
		aggregate.ID(),
		aggregate.Version(),
		events,
	)
}

// FindByID finds aggregate events by its id and rehydrates the aggregate
func (s *Store[T]) FindByID(ctx context.Context, id string) (*T, error) {
	storedEvents, err := s.eventStore.ReadStream(ctx, id)
	if err != nil {
		return nil, err
	}

	var events []Event

	for _, evt := range storedEvents {
		events = append(events, Event{
			ID:                 evt.ID,
			E:                  evt,
			OccurredOn:         evt.OccurredOn,
			CausationEventID:   evt.CausationEventID,
			CorrelationEventID: evt.CorrelationEventID,
			Meta:               evt.Meta,
		})
	}

	var acc T

	acc.Rehydrate(&acc, events...)

	return &acc, nil
}
