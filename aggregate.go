package eventstore

import (
	"fmt"
	"reflect"
)

// AggregateRoot represents reusable DDD Event Sourcing friendly Aggregate
// base type which provides helpers for easy aggregate intialization and
// event handler execution
type AggregateRoot struct {
	version      int
	domainEvents []interface{}
	aggrPtr      reflect.Value
}

// Version returns current version of the aggregate
func (a *AggregateRoot) Version() int { return a.version }

// Events returns unommited domain events
func (a *AggregateRoot) Events() []interface{} {
	if a.domainEvents == nil {
		return []interface{}{}
	}

	return a.domainEvents
}

// Init is used to initialize aggregate (store pointer to the derived type)
// and/or initialize it with provided events (execute all event handlers)
func (a *AggregateRoot) Init(aggrPtr interface{}, evts ...interface{}) error {
	a.aggrPtr = reflect.ValueOf(aggrPtr)

	if a.aggrPtr.Kind() != reflect.Ptr {
		return fmt.Errorf("aggrPtr needs to be a pointer")
	}

	var err error

	for _, evt := range evts {
		err = a.mutate(evt)
		if err != nil {
			return err
		}

		a.version++
	}

	return nil
}

// Apply mutates aggregate (calls respective event handle) and
// appends event to internal slice so they can be retrieved with Events method
// In order for Apply to work the derived aggregate struct needs to implement
// an event handler method for all events it produces eg:
//
// If it produces event of type: SomethingImportantHappened
// Derived aggregate should have the following method implemented:
// func (a *Aggr) OnSomethingImportantHappened(e SomethingImportantHappened) error
func (a *AggregateRoot) Apply(evts ...interface{}) error {
	for _, evt := range evts {
		err := a.mutate(evt)
		if err != nil {
			return err
		}

		a.appendEvent(evt)
	}

	return nil
}

func (a *AggregateRoot) mutate(evt interface{}) error {
	ev := reflect.TypeOf(evt)

	hname := fmt.Sprintf("On%s", ev.Name())

	h := a.aggrPtr.MethodByName(hname)

	if !h.IsValid() {
		return fmt.Errorf("missing aggregate event handler method: %s", hname)
	}

	results := h.Call([]reflect.Value{
		reflect.ValueOf(evt),
	})

	if len(results) == 0 {
		return fmt.Errorf("event handler should return an error")
	}

	if err, ok := results[0].Interface().(error); ok {
		return err
	}

	return nil
}

func (a *AggregateRoot) appendEvent(evt interface{}) {
	a.domainEvents = append(a.domainEvents, evt)
}
