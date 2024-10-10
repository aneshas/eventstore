package aggregate

import (
	"context"
	"errors"
	"github.com/aneshas/eventstore"
)

// ErrAggregateNotFound is returned when aggregate is not found
var ErrAggregateNotFound = errors.New("aggregate not found")

type metaKey struct{}

type correlationIDKey struct{}

type causationIDKey struct{}

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
	var (
		events        []eventstore.EventToStore
		meta          map[string]string
		correlationID string
		causationID   string
	)

	if v, ok := ctx.Value(metaKey{}).(map[string]string); ok {
		meta = v
	}

	if v, ok := ctx.Value(correlationIDKey{}).(string); ok {
		correlationID = v
	}

	if v, ok := ctx.Value(causationIDKey{}).(string); ok {
		causationID = v
	}

	for _, evt := range aggregate.Events() {
		events = append(events, eventstore.EventToStore{
			Event:      evt.E,
			ID:         evt.ID,
			OccurredOn: evt.OccurredOn,

			// Optional
			CausationEventID:   causationID,
			CorrelationEventID: correlationID,
			Meta:               meta,
		})
	}

	return s.eventStore.AppendStream(
		ctx,
		aggregate.ID(),
		aggregate.Version(),
		events,
	)
}

// ByID finds aggregate events by its stream id and rehydrates the aggregate
func (s *Store[T]) ByID(ctx context.Context, id string, root T) error {
	storedEvents, err := s.eventStore.ReadStream(ctx, id)
	if err != nil {
		if errors.Is(err, eventstore.ErrStreamNotFound) {
			return ErrAggregateNotFound
		}

		return err
	}

	var events []Event

	for _, evt := range storedEvents {
		events = append(events, Event{
			ID:                 evt.ID,
			E:                  evt.Event,
			OccurredOn:         evt.OccurredOn,
			CausationEventID:   evt.CausationEventID,
			CorrelationEventID: evt.CorrelationEventID,
			Meta:               evt.Meta,
		})
	}

	root.Rehydrate(root, events...)

	return nil
}

// CtxWithMeta returns new context with meta data
func CtxWithMeta(ctx context.Context, meta map[string]string) context.Context {
	return context.WithValue(ctx, metaKey{}, meta)
}

// CtxWithCorrelationID returns new context with correlation ID
func CtxWithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDKey{}, id)
}

// CtxWithCausationID returns new context with causation ID
func CtxWithCausationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, causationIDKey{}, id)
}
