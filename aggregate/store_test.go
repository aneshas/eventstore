package aggregate_test

import (
	"context"
	"github.com/aneshas/eventstore"
	"github.com/aneshas/eventstore/aggregate"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type eventStore struct {
	eventsToStore []eventstore.EventToStore
	id            string
	ctx           context.Context
	version       int

	storedEvents []eventstore.StoredEvent

	wantErr error
}

// AppendStream appends events to the stream
func (e *eventStore) AppendStream(ctx context.Context, id string, version int, events []eventstore.EventToStore) error {
	for i := range events {
		events[i].OccurredOn = time.Time{}
	}

	e.eventsToStore = events
	e.id = id
	e.version = version
	e.ctx = ctx

	return nil
}

// ReadStream reads events from the stream
func (e *eventStore) ReadStream(_ context.Context, _ string) ([]eventstore.StoredEvent, error) {
	if e.wantErr != nil {
		return nil, e.wantErr
	}

	return e.storedEvents, nil
}

type fooEvent struct {
	Foo string
}

// ID represents an ID
type ID string

func (id ID) String() string {
	return string(id)
}

type foo struct {
	aggregate.Root[ID]

	Balance int
}

func (f *foo) doStuff() {
	f.Apply(
		fooEvent{
			Foo: "foo-1",
		},
		fooEvent{
			Foo: "foo-2",
		},
	)
}

// OnFooEvent handler
func (f *foo) OnfooEvent(evt fooEvent) {
	f.SetID(ID(evt.Foo))
}

func TestShould_Save_Aggregate_Events(t *testing.T) {
	var es eventStore

	store := aggregate.NewStore[*foo](&es)

	meta := map[string]string{
		"foo": "bar",
	}

	ctx := aggregate.CtxWithMeta(context.Background(), meta)
	ctx = aggregate.CtxWithCausationID(ctx, "some-causation-event-id")
	ctx = aggregate.CtxWithCorrelationID(ctx, "some-correlation-event-id")

	var f foo

	f.Rehydrate(&f)
	f.doStuff()

	err := store.Save(ctx, &f)

	assert.NoError(t, err)

	assert.Equal(t, "some-causation-event-id", es.eventsToStore[0].CausationEventID)
	assert.Equal(t, "some-correlation-event-id", es.eventsToStore[0].CorrelationEventID)
	assert.Equal(t, meta, es.eventsToStore[0].Meta)

	assert.Equal(t, ctx, es.ctx)
	assert.Equal(t, 0, es.version)
	assert.Equal(t, "foo-2", es.id)

	events := f.Events()

	assert.Equal(t, []eventstore.EventToStore{
		{
			Event: fooEvent{
				Foo: "foo-1",
			},
			ID:                 events[0].ID,
			CausationEventID:   "some-causation-event-id",
			CorrelationEventID: "some-correlation-event-id",
			Meta:               meta,
			OccurredOn:         time.Time{},
		},
		{
			Event: fooEvent{
				Foo: "foo-2",
			},
			ID:                 events[1].ID,
			CausationEventID:   "some-causation-event-id",
			CorrelationEventID: "some-correlation-event-id",
			Meta:               meta,
			OccurredOn:         time.Time{},
		},
	}, es.eventsToStore)

}

func TestShould_Return_AggregateNotFound_Error_If_No_Events(t *testing.T) {
	var es eventStore

	es.wantErr = eventstore.ErrStreamNotFound

	var f foo

	store := aggregate.NewStore[*foo](&es)

	err := store.ByID(context.Background(), "", &f)

	assert.ErrorIs(t, err, aggregate.ErrAggregateNotFound)
}

func TestShould_Rehydrate_Aggregate(t *testing.T) {
	var es eventStore

	var f foo

	store := aggregate.NewStore[*foo](&es)

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
			StreamID:           "foo-1",
			StreamVersion:      1,
			OccurredOn:         time.Time{},
		},
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
			StreamID:           "foo-2",
			StreamVersion:      1,
			OccurredOn:         time.Time{},
		},
	}

	err := store.ByID(context.Background(), "", &f)

	assert.NoError(t, err)
	assert.Equal(t, "foo-1", f.ID())
	assert.Equal(t, 2, f.Version())
	assert.Len(t, f.Events(), 0)
}
