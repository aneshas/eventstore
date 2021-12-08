package eventstore_test

import (
	"context"
	"flag"
	"os"
	"reflect"
	"testing"

	"github.com/aneshas/goddd/eventstore"
)

var integration = flag.Bool("integration", false, "perform integration tests")

type SomeEvent struct {
	UserID string
}

func TestShouldReadAppendedEvents(t *testing.T) {
	if !*integration {
		return
	}

	es := eventStore(t)

	evts := []interface{}{
		&SomeEvent{
			UserID: "user-1",
		},
		&SomeEvent{
			UserID: "user-2",
		},
		&SomeEvent{
			UserID: "user-2",
		},
	}

	ctx := context.Background()
	stream := "some-stream"
	meta := map[string]string{
		"ip": "127.0.0.1",
	}

	err := es.AppendStream(
		ctx, stream, eventstore.InitialStreamVersion, evts,
		eventstore.WithMetaData(meta),
	)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	got, err := es.ReadStream(ctx, stream)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	for i, evt := range got {
		if !reflect.DeepEqual(evt.Event, evts[i]) ||
			!reflect.DeepEqual(evt.Meta, meta) ||
			evt.Type != "SomeEvent" {

			t.Logf("%#v", evts)
			t.Logf("%#v", got)
			t.Fatal("events not read")
		}
	}
}

func TestShouldWriteToDifferentStreams(t *testing.T) {
	if !*integration {
		return
	}

	es := eventStore(t)

	evts := []interface{}{
		&SomeEvent{
			UserID: "user-1",
		},
		&SomeEvent{
			UserID: "user-2",
		},
		&SomeEvent{
			UserID: "user-2",
		},
	}

	ctx := context.Background()
	streamOne := "some-stream"
	streamTwo := "another-stream"

	err := es.AppendStream(
		ctx, streamOne, eventstore.InitialStreamVersion, evts,
	)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	err = es.AppendStream(
		ctx, streamTwo, eventstore.InitialStreamVersion, evts,
	)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestShouldAppendToExistingStream(t *testing.T) {
	if !*integration {
		return
	}

	es := eventStore(t)

	evts := []interface{}{
		&SomeEvent{
			UserID: "user-1",
		},
		&SomeEvent{
			UserID: "user-2",
		},
		&SomeEvent{
			UserID: "user-2",
		},
	}

	ctx := context.Background()
	stream := "some-stream"

	err := es.AppendStream(
		ctx, stream, eventstore.InitialStreamVersion, evts,
	)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	err = es.AppendStream(
		ctx, stream, 3, evts,
	)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestOptimisticConcurrencyCheckIsPerformed(t *testing.T) {
	if !*integration {
		return
	}

	es := eventStore(t)

	evts := []interface{}{
		&SomeEvent{
			UserID: "user-1",
		},
	}

	ctx := context.Background()
	stream := "some-stream"

	err := es.AppendStream(
		ctx, stream, eventstore.InitialStreamVersion, evts,
	)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	err = es.AppendStream(
		ctx, stream, eventstore.InitialStreamVersion, evts,
	)

	if err != eventstore.ErrConcurrencyCheckFailed {
		t.Fatalf("should have performed optimistic concurrency check")
	}
}

func TestReadStreamWrapsNotFoundError(t *testing.T) {
	if !*integration {
		return
	}

	es := eventStore(t)

	_, err := es.ReadStream(context.Background(), "foo-stream")
	t.Logf("err: %v", err)
	if err != eventstore.ErrStreamNotFound {
		t.Fatal("should return explicit error if stream doesn't exist")
	}
}

func eventStore(t *testing.T) *eventstore.EventStore {
	file, err := os.CreateTemp(os.TempDir(), "estore-data")
	if err != nil {
		t.Fatalf("could not create tem file: %v", err)
	}

	es, err := eventstore.New(file.Name(), eventstore.NewJsonEncoder(SomeEvent{}))
	if err != nil {
		t.Fatal(err)
	}

	return es
}
