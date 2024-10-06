package aggregate

import (
	"fmt"
	"reflect"
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

// Root represents reusable DDD Event Sourcing friendly Aggregate
// base type which provides helpers for easy aggregate initialization and
// event handler execution
type Root[T any] struct {
	ID T

	version      int
	domainEvents []any

	ptr reflect.Value
}

// Rehydrate is used to construct and rehydrate the aggregate from events
func (a *Root[T]) Rehydrate(aggregatePtr any, events ...any) {
	a.ptr = reflect.ValueOf(aggregatePtr)

	if a.ptr.Kind() != reflect.Ptr {
		panic(ErrAggregateRootNotAPointer)
	}

	for _, evt := range events {
		a.mutate(evt)

		a.version++
	}

}

// Version returns current version of the aggregate (incremented every time
// Apply is successfully called)
func (a *Root[T]) Version() int { return a.version }

// Events returns uncommitted domain events (produced by calling Apply)
func (a *Root[T]) Events() []any {
	if a.domainEvents == nil {
		return []any{}
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
// func (a *SomeAggregate) OnSomethingImportantHappened(e SomethingImportantHappened) error
func (a *Root[T]) Apply(events ...any) {
	if !a.ptr.IsValid() {
		panic(ErrAggregateRootNotRehydrated)
	}

	for _, evt := range events {
		a.mutate(evt)

		a.appendEvent(evt)
	}
}

func (a *Root[T]) mutate(evt any) {
	ev := reflect.TypeOf(evt)

	hName := fmt.Sprintf("On%s", ev.Name())

	h := a.ptr.MethodByName(hName)

	if !h.IsValid() {
		panic(ErrMissingAggregateEventHandler)
	}

	h.Call([]reflect.Value{
		reflect.ValueOf(evt),
	})
}

func (a *Root[T]) appendEvent(evt any) {
	a.domainEvents = append(a.domainEvents, evt)
}
