package eventstore_test

import (
	"testing"

	"github.com/aneshas/eventstore"
)

type testAggregate struct {
	eventstore.AggregateRoot

	name  string
	email string
}

func (ta *testAggregate) OnaggrCreated(event aggrCreated) {
	ta.name = event.Name()
	ta.email = event.Email()
}

func (ta *testAggregate) OnnameUpdated(event nameUpdated) {
	ta.name = event.NewName()
}

type aggrCreated struct {
	name  string
	email string
}

func (e *aggrCreated) Name() string  { return e.name }
func (e *aggrCreated) Email() string { return e.email }

type nameUpdated struct {
	newName string
}

func (e *nameUpdated) NewName() string { return e.newName }

func TestApplyEventShouldMutateAggregateAndAddEvent(t *testing.T) {
	aggr := testAggregate{}

	aggr.Init(&aggr)
	aggr.ApplyEvent(aggrCreated{"john", "john@email.com"})
	aggr.ApplyEvent(nameUpdated{"max"})

	events := aggr.Events()

	if len(events) != 2 {
		t.Errorf("event count should be 2")
	}

	if aggr.name != "max" || aggr.email != "john@email.com" {
		t.Errorf("aggregate not mutated")
	}
}

func TestShouldInitAggregate(t *testing.T) {
	aggr := testAggregate{}

	aggr.Init(
		&aggr,
		aggrCreated{"john", "john@email.com"},
		nameUpdated{"max"},
	)

	aggr.ApplyEvent(nameUpdated{"jane"})

	if aggr.name != "jane" || aggr.email != "john@email.com" {
		t.Errorf("aggregate not mutated")
	}
}
