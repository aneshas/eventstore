package goddd

import (
	"reflect"
)

// // DomainID is an interface of a domain identity
// type DomainID interface {
// 	fmt.Stringer
// }

// // AggregateRoot represents DDD Aggregate interface
// type AggregateRoot interface {
// 	// Events should return uncommited domain events
// 	Events() []interface{}

// 	// Version should return aggregate version
// 	// used for optimistic concurrency
// 	Version() int
// }

// AggregateRoot represents reusable DDD Aggregate implementation
type AggregateRoot struct {
	version      int
	DomainEvents []interface{}
	aggrPtr      interface{}
}

// Version returns aggregate version
func (a *AggregateRoot) Version() int { return a.version }

// TODO - Fix comments

// Events returns aggregate domain events
func (a *AggregateRoot) Events() []interface{} {
	if a.DomainEvents == nil {
		return []interface{}{}
	}

	return a.DomainEvents
}

func (a *AggregateRoot) Init(aggrPtr interface{}, evts ...interface{}) {
	a.aggrPtr = aggrPtr

	for _, evt := range evts {
		// TODO - Panic if mutate errors out - don't !!!
		a.mutate(evt)
		a.version++
	}
}

// ApplyEvent mutates aggregate and appends event to domain events
func (a *AggregateRoot) ApplyEvent(evt interface{}) {
	// TODO Apply should return error in case mutate fails
	a.mutate(evt)
	a.AppendEvent(evt)
}

func (a *AggregateRoot) mutate(evt interface{}) {
	v := reflect.ValueOf(a.aggrPtr)
	ev := reflect.TypeOf(evt)

	handle := v.MethodByName("On" + ev.Elem().Name())

	// TODO better errors

	handle.Call([]reflect.Value{
		reflect.ValueOf(evt),
	})
}

// AppendEvent would be used if we only made use of
// domain events without event sourcing
func (a *AggregateRoot) AppendEvent(evt interface{}) {
	a.DomainEvents = append(a.DomainEvents, evt)
}
