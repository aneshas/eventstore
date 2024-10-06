package aggregate_test

import (
	"errors"
	"github.com/aneshas/eventstore/aggregate"
	"testing"
)

type created struct {
	name  string
	email string
}

type nameUpdated struct {
	newName string
}

type wrongHandler struct{}
type missingHandler struct{}

type id string

type testAggregate struct {
	aggregate.Root[id]

	name  string
	email string
}

func (ta *testAggregate) Oncreated(event created) {
	ta.name = event.Name()
	ta.email = event.Email()
}

func (ta *testAggregate) OnnameUpdated(event nameUpdated) {
	ta.name = event.NewName()
}

var errTest = errors.New("an error")

func (ta *testAggregate) OnwrongHandler(event wrongHandler, n int) {
}

func (e *created) Name() string  { return e.name }
func (e *created) Email() string { return e.email }

func (e *nameUpdated) NewName() string { return e.newName }

func TestApplyEventShouldMutateAggregateAndAddEvent(t *testing.T) {
	var a testAggregate

	a.Rehydrate(&a)

	a.Apply(created{"john", "john@email.com"})
	a.Apply(nameUpdated{"max"})

	events := a.Events()

	if len(events) != 2 {
		t.Errorf("event count should be 2")
	}

	if a.name != "max" || a.email != "john@email.com" {
		t.Errorf("aggregate not mutated")
	}
}

func TestShouldInitAggregate(t *testing.T) {
	var a testAggregate

	a.Rehydrate(
		&a,
		created{"john", "john@email.com"},
		nameUpdated{"max"},
	)

	a.Apply(nameUpdated{"jane"})

	if a.name != "jane" || a.email != "john@email.com" {
		t.Errorf("aggregate not mutated")
	}
}

func TestShouldPanicOnApplyWithNoRehydrate(t *testing.T) {
	defer func() {
		r := recover()

		if r == nil {
			t.Errorf("should")
		}

		err, ok := r.(error)

		if !ok {
			t.Errorf("should panic with error")
		}

		if !errors.Is(err, aggregate.ErrAggregateRootNotRehydrated) {
			t.Errorf("should panic with not rehydrated error")
		}
	}()

	var a testAggregate

	a.Apply(missingHandler{})
}

func TestShouldPanicOnMissingHandler(t *testing.T) {
	defer func() {
		r := recover()

		if r == nil {
			t.Errorf("should")
		}

		err, ok := r.(error)

		if !ok {
			t.Errorf("should panic with error")
		}

		if !errors.Is(err, aggregate.ErrMissingAggregateEventHandler) {
			t.Errorf("should panic with missing handler error")
		}
	}()

	var a testAggregate

	a.Rehydrate(&a)

	a.Apply(missingHandler{})
}

func TestShouldAcceptOnlyPointerOnRehydration(t *testing.T) {
	defer func() {
		r := recover()

		if r == nil {
			t.Errorf("should panic")
		}

		err, ok := r.(error)

		if !ok {
			t.Errorf("should panic with error")
		}

		if !errors.Is(err, aggregate.ErrAggregateRootNotAPointer) {
			t.Errorf("should panic with pointer error")
		}
	}()

	var a testAggregate

	a.Rehydrate(a)
}
