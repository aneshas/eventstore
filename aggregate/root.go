package aggregate

import (
	"fmt"
	"github.com/google/uuid"
	"reflect"
	"time"
)

var (
	// ErrMissingAggregateEventHandler is returned when aggregate event handler is missing
	// On{EventName} method
	ErrMissingAggregateEventHandler = fmt.Errorf("missing aggregate event handler")

	// ErrAggregateRootNotAPointer is returned when supplied aggregate root is not a pointer
	ErrAggregateRootNotAPointer = fmt.Errorf("aggregate needs to be a pointer")

	// ErrAggregateRootNotRehydrated is returned when aggregate is not rehydrated (with Rehydrate method)
	ErrAggregateRootNotRehydrated = fmt.Errorf("aggregate needs to be rehydrated")
)

// Rooter represents an aggregate root interface
type Rooter interface {
	StringID() string
	Events() []Event
	Version() int
	Rehydrate(acc any, events ...Event)
	FirstEventID() string
	LastEventID() string
}

// Root represents reusable DDD Event Sourcing friendly Aggregate
// base type which provides helpers for easy aggregate initialization and
// event handler execution
type Root[T fmt.Stringer] struct {
	ID T

	version      int
	domainEvents []Event

	firstEventID string
	lastEventID  string

	ptr reflect.Value
}

// LastEventID returns last event ID
func (a *Root[T]) LastEventID() string {
	return a.lastEventID
}

// FirstEventID returns first event ID
func (a *Root[T]) FirstEventID() string {
	return a.firstEventID
}

// StringID returns aggregate ID string
func (a *Root[T]) StringID() string {
	return a.ID.String()
}

// Rehydrate is used to construct and rehydrate the aggregate from events
func (a *Root[T]) Rehydrate(aggregatePtr any, events ...Event) {
	a.ptr = reflect.ValueOf(aggregatePtr)

	if a.ptr.Kind() != reflect.Ptr {
		panic(ErrAggregateRootNotAPointer)
	}

	for _, evt := range events {
		a.mutate(evt)
		a.lastEventID = evt.ID

		a.version++
	}
}

// Version returns current version of the aggregate (incremented every time
// Apply is successfully called)
func (a *Root[T]) Version() int { return a.version }

// Events returns uncommitted domain events (produced by calling Apply)
func (a *Root[T]) Events() []Event {
	if a.domainEvents == nil {
		return []Event{}
	}

	return a.domainEvents
}

// Apply mutates aggregate (calls respective event handle) and
// appends event to internal slice, so that they can be retrieved with Events method
// In order for Apply to work the derived aggregate struct needs to implement
// an event handler method for all events it produces eg:
//
// If it produces event of type: SomethingImportantHappened
// Derived aggregate should have the following method implemented:
// func (a *SomeAggregate) OnSomethingImportantHappened(event SomethingImportantHappened) error
// or
// func (a *SomeAggregate) OnSomethingImportantHappened(event SomethingImportantHappened, extra aggregate.Event) error
func (a *Root[T]) Apply(events ...any) {
	if !a.ptr.IsValid() {
		panic(ErrAggregateRootNotRehydrated)
	}

	for _, evt := range events {
		e := Event{
			ID:         uuid.Must(uuid.NewV7()).String(),
			E:          evt,
			OccurredOn: time.Now().UTC(),
		}

		a.mutate(e)
		a.appendEvent(e)
	}
}

// ApplyWithID applies single event and mutates aggregate (calls respective event handle) and sets event ID explicitly.
// See Apply for more details.
func (a *Root[T]) ApplyWithID(eventID string, event any) {
	if !a.ptr.IsValid() {
		panic(ErrAggregateRootNotRehydrated)
	}

	e := Event{
		ID:         eventID,
		E:          event,
		OccurredOn: time.Now().UTC(),
	}

	a.mutate(e)
	a.appendEvent(e)
}

func (a *Root[T]) mutate(evt Event) {
	ev := reflect.TypeOf(evt.E)

	hName := fmt.Sprintf("On%s", ev.Name())

	h := a.ptr.MethodByName(hName)

	if !h.IsValid() {
		panic(ErrMissingAggregateEventHandler)
	}

	if a.firstEventID == "" {
		a.firstEventID = evt.ID
	}

	if h.Type().NumIn() == 2 {
		h.Call([]reflect.Value{
			reflect.ValueOf(evt.E),
			reflect.ValueOf(evt),
		})

		return
	}

	h.Call([]reflect.Value{
		reflect.ValueOf(evt.E),
	})
}

func (a *Root[T]) appendEvent(evt Event) {
	a.domainEvents = append(a.domainEvents, evt)
}
