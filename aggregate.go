package goddd

import (
	"fmt"
	"reflect"
)

// DomainID is an interface of a domain identity
type DomainID interface {
	fmt.Stringer
}

// DomainEvent represents DDD domain event
type DomainEvent interface {
	// TODO - Timestamp?
}

// AggregateRoot represents DDD Aggregate interface
type AggregateRoot interface {
	// ID should return aggregate identity
	ID() DomainID

	// Events should return uncommited domain events
	Events() []DomainEvent

	// Version should return aggregates version
	// used for optimistic concurrency
	Version() int
}

// BaseAggregateRoot represents reusable DDD Aggregate implementation
type BaseAggregateRoot struct {
	version      int
	domainEvents []DomainEvent
}

// Version returns BaseAggregates version
func (a *BaseAggregateRoot) Version() int { return a.version }

// Events returns BaseAggregate domain events
func (a *BaseAggregateRoot) Events() []DomainEvent {
	if a.domainEvents == nil {
		return []DomainEvent{}
	}

	return a.domainEvents
}

func (a *BaseAggregateRoot) Init(aggrPtr AggregateRoot, evts []DomainEvent) {
	for _, evt := range evts {
		// TODO - Panic if mutate errors out
		a.mutate(aggrPtr, evt)
		a.version++
	}
}

// ApplyEvent mutates aggregate and appends event to domain events
func (a *BaseAggregateRoot) ApplyEvent(aggrPtr AggregateRoot, evt DomainEvent) {
	// TODO Apply should return error in case mutate fails
	a.mutate(aggrPtr, evt)
	a.AppendEvent(evt)
}

func (a *BaseAggregateRoot) mutate(aggrPtr AggregateRoot, evt DomainEvent) {
	v := reflect.ValueOf(aggrPtr)
	ev := reflect.TypeOf(evt)

	handle := v.MethodByName("On" + ev.Elem().Name())

	// TODO better errors

	handle.Call([]reflect.Value{
		reflect.ValueOf(evt),
	})
}

// AppendEvent would be used if we only made use of
// domain events without event sourcing
func (a *BaseAggregateRoot) AppendEvent(evt DomainEvent) {
	a.domainEvents = append(a.domainEvents, evt)
}
