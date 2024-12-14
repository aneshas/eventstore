package aggregate_test

import (
	"context"
	"fmt"
	"github.com/aneshas/eventstore"
	"github.com/aneshas/eventstore/aggregate"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestShould_Load_And_Persist_Aggregate(t *testing.T) {
	var es eventStore

	store := aggregate.NewStore[*foo](&es)

	aggrID := "foo-1"

	es.storedEvents = []eventstore.StoredEvent{
		{
			Event: fooEvent{
				Foo: "foo-1",
			},
			Meta:               nil,
			ID:                 "event-id-1",
			Sequence:           1,
			Type:               "fooEvent",
			CausationEventID:   nil,
			CorrelationEventID: nil,
			StreamID:           aggrID,
			StreamVersion:      1,
			OccurredOn:         time.Time{},
		},
	}

	exec := aggregate.NewExecutor(store)

	var f foo

	f.ID = "foo-1"

	err := exec(context.Background(), &f, func(ctx context.Context) error {
		f.doMoreStuff()

		return nil
	})

	assert.NoError(t, err)

	assert.Equal(t, es.storedEvents[0].Event.(fooEvent).Foo, aggrID)
}

func TestShould_Should_Report_Exec_Error(t *testing.T) {
	var es eventStore

	store := aggregate.NewStore[*foo](&es)

	aggrID := "foo-1"

	es.storedEvents = []eventstore.StoredEvent{
		{
			Event: fooEvent{
				Foo: "foo-1",
			},
			Meta:               nil,
			ID:                 "event-id-1",
			Sequence:           1,
			Type:               "fooEvent",
			CausationEventID:   nil,
			CorrelationEventID: nil,
			StreamID:           aggrID,
			StreamVersion:      1,
			OccurredOn:         time.Time{},
		},
	}

	exec := aggregate.NewExecutor(store)

	var f foo

	f.ID = "foo-1"

	wantErr := fmt.Errorf("error")

	err := exec(context.Background(), &f, func(ctx context.Context) error {
		return wantErr
	})

	assert.ErrorIs(t, err, wantErr)
}

func TestShould_Report_AggregateNotFound_Error(t *testing.T) {
	var es eventStore

	store := aggregate.NewStore[*foo](&es)

	exec := aggregate.NewExecutor(store)

	var f foo

	f.ID = "foo-1"

	es.wantErr = eventstore.ErrStreamNotFound

	err := exec(context.Background(), &f, func(ctx context.Context) error {
		f.doMoreStuff()

		return nil
	})

	assert.ErrorIs(t, err, aggregate.ErrAggregateNotFound)
}
