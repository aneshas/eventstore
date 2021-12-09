package eventstore

// TODO - Consider moving this to eventstore package since it is only
// useful in that context
import (
	"reflect"
)

// AggregateRoot represents reusable DDD Aggregate implementation
type AggregateRoot struct {
	version      int
	domainEvents []interface{}
	aggrPtr      interface{}
}

// Version returns aggregate version
func (a *AggregateRoot) Version() int { return a.version }

// TODO - Fix comments

// Events returns aggregate domain events
func (a *AggregateRoot) Events() []interface{} {
	if a.domainEvents == nil {
		return []interface{}{}
	}

	return a.domainEvents
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
	a.appendEvent(evt)
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

func (a *AggregateRoot) appendEvent(evt interface{}) {
	a.domainEvents = append(a.domainEvents, evt)
}
