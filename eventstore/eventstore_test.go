package eventstore_test

import (
	"context"
	"errors"
	"flag"
	"io"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/aneshas/goddd/eventstore"
)

var integration = flag.Bool("integration", true, "perform integration tests")

type SomeEvent struct {
	UserID string
}

func TestShouldReadAppendedEvents(t *testing.T) {
	if !*integration {
		return
	}

	es, cleanup := eventStore(t)

	defer cleanup()

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

			t.Fatal("events not read")
		}
	}
}

func TestShouldWriteToDifferentStreams(t *testing.T) {
	if !*integration {
		return
	}

	es, cleanup := eventStore(t)

	defer cleanup()

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

	es, cleanup := eventStore(t)

	defer cleanup()

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

	es, cleanup := eventStore(t)

	defer cleanup()

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

	es, cleanup := eventStore(t)

	defer cleanup()

	_, err := es.ReadStream(context.Background(), "foo-stream")
	if err != eventstore.ErrStreamNotFound {
		t.Fatal("should return explicit error if stream doesn't exist")
	}
}

func TestReadAllCatchesUpToNewEvents(t *testing.T) {
	if !*integration {
		return
	}

	es, cleanup := eventStore(t)

	defer cleanup()

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

	err := es.AppendStream(ctx, "stream-one", eventstore.InitialStreamVersion, evts)
	if err != nil {
		t.Fatal(err)
	}

	sub, _ := es.ReadAll(ctx)

	defer sub.Close()

	got := readAllSub(t, sub)

	if len(got) != 3 {
		t.Fatal("should have read 3 events")
	}

	evtsTwo := []interface{}{
		&SomeEvent{
			UserID: "user-1",
		},
		&SomeEvent{
			UserID: "user-2",
		},
		&SomeEvent{
			UserID: "user-2",
		},
		&SomeEvent{
			UserID: "user-2",
		},
	}

	err = es.AppendStream(ctx, "stream-two", eventstore.InitialStreamVersion, evtsTwo)
	if err != nil {
		t.Fatal(err)
	}

	got = readAllSub(t, sub)

	// TODO - Check event details

	if len(got) != 4 {
		t.Fatal("should have read 4 events")
	}
}

func readAllSub(t *testing.T, sub eventstore.Subscription) []eventstore.EventData {
	var got []eventstore.EventData

outer:
	for {
		select {
		case data := <-sub.EventData:
			got = append(got, data)

		case err := <-sub.Err:
			if err != nil {
				if errors.Is(err, io.EOF) {
					if len(got) == 0 {
						break
					}
					break outer
				}

				t.Fatal(err)
			}
		}
	}

	return got
}

func TestReadAllCancelsSubscriptionOnContextCancel(t *testing.T) {
	if !*integration {
		return
	}

	es, cleanup := eventStore(t)

	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	_ = cancel

	sub, _ := es.ReadAll(ctx)

	timeout := time.After(2 * time.Second)

	for {
		select {
		case <-timeout:
			t.Fatal("subscription should have been closed")
		case err := <-sub.Err:
			if errors.Is(err, io.EOF) {
				break
			}

			return
		}
	}
}

func TestReadAllCancelsSubscriptionWithClose(t *testing.T) {
	if !*integration {
		return
	}

	es, cleanup := eventStore(t)

	defer cleanup()

	sub, _ := es.ReadAll(context.Background())

	go func() {
		time.Sleep(time.Second)

		sub.Close()
	}()

	timeout := time.After(2 * time.Second)

	for {
		select {
		case <-timeout:
			t.Fatal("subscription should have been closed")
		case err := <-sub.Err:
			if errors.Is(err, io.EOF) {
				break
			}

			if !errors.Is(err, eventstore.ErrSubscriptionClosedByClient) {
				t.Fatal("incorrect subscription cancel error")
			}

			return
		}
	}
}

func eventStore(t *testing.T) (*eventstore.EventStore, func()) {
	file, err := os.CreateTemp(os.TempDir(), "es-db-*")
	if err != nil {
		t.Fatalf("could not create tem file: %v", err)
	}

	es, err := eventstore.New(file.Name(), eventstore.NewJsonEncoder(SomeEvent{}))
	if err != nil {
		t.Fatalf("error creating es: %v", err)
	}

	err = file.Close()
	if err != nil {
		t.Fatal(err)
	}

	return es, func() {
		err := es.Close()
		if err != nil {
			t.Fatal(err)
		}

		err = os.Remove(file.Name())
		if err != nil {
			t.Fatal(err)
		}
	}
}
