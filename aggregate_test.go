package eventstore_test

import (
	"errors"
	"testing"

	"github.com/aneshas/eventstore"
)

type aggrCreated struct {
	name  string
	email string
}

type nameUpdated struct {
	newName string
}

type erroredOut struct{}
type wrongHandler struct{}
type missingHandler struct{}

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

var errTest = errors.New("an error")

func (ta *testAggregate) OnwrongHandler(event wrongHandler, n int) {
}

func (e *aggrCreated) Name() string  { return e.name }
func (e *aggrCreated) Email() string { return e.email }

func (e *nameUpdated) NewName() string { return e.newName }

func TestApplyEventShouldMutateAggregateAndAddEvent(t *testing.T) {
	aggr := testAggregate{}

	aggr.Init(&aggr)
	aggr.Apply(aggrCreated{"john", "john@email.com"})
	aggr.Apply(nameUpdated{"max"})

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

	aggr.Apply(nameUpdated{"jane"})

	if aggr.name != "jane" || aggr.email != "john@email.com" {
		t.Errorf("aggregate not mutated")
	}
}

func TestShouldErrorOutOnMissingHandler(t *testing.T) {
	aggr := testAggregate{}

	aggr.Init(&aggr)

	err := aggr.Apply(missingHandler{})
	if err == nil {
		t.Fatal("should error out on missing event handler")
	}
}

func TestInitAcceptsOnlyPointerValues(t *testing.T) {
	aggr := testAggregate{}

	err := aggr.Init(aggr)
	if err == nil {
		t.Fatal("should error out if ptr value not provided")
	}
}
