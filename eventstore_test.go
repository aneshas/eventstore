package eventstore_test

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/aneshas/eventstore"
)

var integration = flag.Bool("integration", false, "perform integration tests")

type SomeEvent struct {
	UserID string
}

func TestShouldReadAppendedEvents(t *testing.T) {
	if !*integration {
		t.Skip("skipping integration tests")
	}

	es, cleanup := eventStore(t)

	defer cleanup()

	evts := []interface{}{
		SomeEvent{
			UserID: "user-1",
		},
		SomeEvent{
			UserID: "user-2",
		},
		SomeEvent{
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
		t.Skip("skipping integration tests")
	}

	es, cleanup := eventStore(t)

	defer cleanup()

	evts := []interface{}{
		SomeEvent{
			UserID: "user-1",
		},
		SomeEvent{
			UserID: "user-2",
		},
		SomeEvent{
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
		t.Skip("skipping integration tests")
	}

	es, cleanup := eventStore(t)

	defer cleanup()

	evts := []interface{}{
		SomeEvent{
			UserID: "user-1",
		},
		SomeEvent{
			UserID: "user-2",
		},
		SomeEvent{
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
		t.Skip("skipping integration tests")
	}

	es, cleanup := eventStore(t)

	defer cleanup()

	evts := []interface{}{
		SomeEvent{
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
		t.Skip("skipping integration tests")
	}

	es, cleanup := eventStore(t)

	defer cleanup()

	_, err := es.ReadStream(context.Background(), "foo-stream")
	if err != eventstore.ErrStreamNotFound {
		t.Fatal("should return explicit error if stream doesn't exist")
	}
}

func TestSubscribeAllWithOffsetCatchesUpToNewEvents(t *testing.T) {
	if !*integration {
		t.Skip("skipping integration tests")
	}

	es, cleanup := eventStore(t)

	defer cleanup()

	evts := []interface{}{
		SomeEvent{
			UserID: "user-1",
		},
		SomeEvent{
			UserID: "user-2",
		},
		SomeEvent{
			UserID: "user-2",
		},
	}

	ctx := context.Background()

	err := es.AppendStream(ctx, "stream-one", eventstore.InitialStreamVersion, evts)
	if err != nil {
		t.Fatal(err)
	}

	sub, err := es.SubscribeAll(
		ctx,
		eventstore.WithOffset(1),
		eventstore.WithPollInterval(50*time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}

	defer sub.Close()

	got := readAllSub(t, sub, 2)

	if len(got) != 2 {
		t.Fatalf("should have read 2 events. actual: %d", len(got))
	}

	evtsTwo := []interface{}{
		SomeEvent{
			UserID: "user-1",
		},
		SomeEvent{
			UserID: "user-2",
		},
		SomeEvent{
			UserID: "user-2",
		},
		SomeEvent{
			UserID: "user-2",
		},
	}

	err = es.AppendStream(ctx, "stream-two", eventstore.InitialStreamVersion, evtsTwo)
	if err != nil {
		t.Fatal(err)
	}

	got = readAllSub(t, sub, 4)

	if len(got) != 4 {
		t.Fatalf("should have read 4 events. actual: %d", len(got))
	}
}

func readAllSub(t *testing.T, sub eventstore.Subscription, expect int) []eventstore.EventData {
	var got []eventstore.EventData

outer:
	for {
		select {
		case data := <-sub.EventData:
			got = append(got, data)

		case err := <-sub.Err:
			if err != nil {
				if errors.Is(err, io.EOF) {
					if len(got) < expect {
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

func TestReadAllShouldReadAllEvents(t *testing.T) {
	if !*integration {
		t.Skip("skipping integration tests")
	}

	es, cleanup := eventStore(t)

	defer cleanup()

	evts := []interface{}{
		SomeEvent{
			UserID: "user-1",
		},
		SomeEvent{
			UserID: "user-2",
		},
		SomeEvent{
			UserID: "user-3",
		},
	}

	ctx := context.Background()

	err := es.AppendStream(ctx, "stream-one", eventstore.InitialStreamVersion, evts)
	if err != nil {
		t.Fatal(err)
	}

	data, err := es.ReadAll(ctx)
	if err != nil {
		t.Fatal(err)
	}

	got := []interface{}{
		data[0].Event.(SomeEvent),
		data[1].Event.(SomeEvent),
		data[2].Event.(SomeEvent),
	}

	if !reflect.DeepEqual(evts, got) {
		t.Fatal("all events should have been read")
	}
}

func TestSubscribeAllCancelsSubscriptionOnContextCancel(t *testing.T) {
	if !*integration {
		t.Skip("skipping integration tests")
	}

	es, cleanup := eventStore(t)

	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	_ = cancel

	sub, _ := es.SubscribeAll(ctx)

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

func TestSubscribeAllCancelsSubscriptionWithClose(t *testing.T) {
	if !*integration {
		t.Skip("skipping integration tests")
	}

	es, cleanup := eventStore(t)

	defer cleanup()

	sub, _ := es.SubscribeAll(context.Background())

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

type enc struct {
	encode func(interface{}) (*eventstore.EncodedEvt, error)
	decode func(*eventstore.EncodedEvt) (interface{}, error)
}

func (e enc) Encode(evt interface{}) (*eventstore.EncodedEvt, error) {
	return e.encode(evt)
}

func (e enc) Decode(evt *eventstore.EncodedEvt) (interface{}, error) {
	return e.decode(evt)
}

func TestEncoderEncodeErrorsPropagated(t *testing.T) {
	if !*integration {
		t.Skip("skipping integration tests")
	}

	var anErr = fmt.Errorf("an error occurred")

	e := enc{
		encode: func(i interface{}) (*eventstore.EncodedEvt, error) { return nil, anErr },
	}

	es, cleanup := eventStoreWithDec(t, e)

	defer cleanup()

	err := es.AppendStream(
		context.Background(),
		"stream",
		eventstore.InitialStreamVersion,
		[]interface{}{
			SomeEvent{
				UserID: "123",
			},
		},
	)

	if !errors.Is(err, anErr) {
		t.Fatal("error should have been propagated")
	}
}

func TestEncoderDecodeErrorsPropagated(t *testing.T) {
	if !*integration {
		t.Skip("skipping integration tests")
	}

	var anErr = fmt.Errorf("an error occurred")

	e := enc{
		encode: func(i interface{}) (*eventstore.EncodedEvt, error) {
			return &eventstore.EncodedEvt{
				Data: "malformed-json",
				Type: "foo",
			}, nil
		},
		decode: func(ee *eventstore.EncodedEvt) (interface{}, error) {
			return nil, anErr
		},
	}

	es, cleanup := eventStoreWithDec(t, e)

	defer cleanup()

	err := es.AppendStream(
		context.Background(),
		"stream",
		eventstore.InitialStreamVersion,
		[]interface{}{
			SomeEvent{
				UserID: "123",
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = es.ReadStream(context.Background(), "stream")

	if !errors.Is(err, anErr) {
		t.Fatal("error should have been propagated")
	}
}

func TestEncoderDecodeErrorsPropagatedOnSubscribeAll(t *testing.T) {
	if !*integration {
		t.Skip("skipping integration tests")
	}

	var anErr = fmt.Errorf("an error occurred")

	e := enc{
		encode: func(i interface{}) (*eventstore.EncodedEvt, error) {
			return &eventstore.EncodedEvt{
				Data: "malformed-json",
				Type: "foo",
			}, nil
		},
		decode: func(ee *eventstore.EncodedEvt) (interface{}, error) {
			return nil, anErr
		},
	}

	es, cleanup := eventStoreWithDec(t, e)

	defer cleanup()

	err := es.AppendStream(
		context.Background(),
		"stream",
		eventstore.InitialStreamVersion,
		[]interface{}{
			SomeEvent{
				UserID: "123",
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	sub, err := es.SubscribeAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	defer sub.Close()

	err = <-sub.Err

	if !errors.Is(err, anErr) {
		t.Fatal("error should have been propagated")
	}
}
func TestNewEncoderMustBeProvided(t *testing.T) {
	_, err := eventstore.New("foo", nil)
	if err == nil {
		t.Fatal("encoder must be provided")
	}
}

func TestAppendStreamValidation(t *testing.T) {
	es := eventstore.EventStore{}

	cases := []struct {
		stream string
		ver    int
		evts   []interface{}
	}{
		{
			stream: "",
			ver:    0,
			evts: []interface{}{
				SomeEvent{
					UserID: "user-123",
				},
			},
		},
		{
			stream: "s",
			ver:    -1,
			evts: []interface{}{
				SomeEvent{
					UserID: "user-123",
				},
			},
		},
		{
			stream: "stream",
			ver:    0,
			evts:   nil,
		},

		{
			stream: "stream",
			ver:    0,
			evts:   []interface{}{},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			err := es.AppendStream(context.Background(), tc.stream, tc.ver, tc.evts)
			if err == nil {
				t.Fatal("validation error should have happened")
			}
		})
	}
}

func TestSubscribeAllMinimumBatchSize(t *testing.T) {
	es := eventstore.EventStore{}

	_, err := es.SubscribeAll(context.Background(), eventstore.WithBatchSize(-1))
	if err == nil {
		t.Fatal("minimum batch size should have been validated")
	}
}

func TestReadAllMinimumBatchSize(t *testing.T) {
	es := eventstore.EventStore{}

	_, err := es.ReadAll(context.Background(), eventstore.WithBatchSize(-1))
	if err == nil {
		t.Fatal("minimum batch size should have been validated")
	}
}

func TestReadStreamValidation(t *testing.T) {
	es := eventstore.EventStore{}

	_, err := es.ReadStream(context.Background(), "")
	if err == nil {
		t.Fatal("stream name should be provided")
	}
}

func eventStore(t *testing.T) (*eventstore.EventStore, func()) {
	return eventStoreWithDec(t, eventstore.NewJSONEncoder(SomeEvent{}))
}

func eventStoreWithDec(t *testing.T, enc eventstore.Encoder) (*eventstore.EventStore, func()) {
	es, err := eventstore.New("file::memory:?cache=shared", enc)
	if err != nil {
		t.Fatalf("error creating es: %v", err)
	}

	return es, func() {
		err := es.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}
