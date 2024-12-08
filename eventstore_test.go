package eventstore_test

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/aneshas/tx/v2"
	"github.com/aneshas/tx/v2/gormtx"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/aneshas/eventstore"
)

var withPG = flag.Bool("withpg", false, "run tests with postgres")

type SomeEvent struct {
	UserID string
}

func TestShouldReadAppendedEvents(t *testing.T) {
	es, cleanup := eventStore(t)

	defer cleanup()

	evts := []any{
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
		ctx, stream, eventstore.InitialStreamVersion, toEventToStore(evts...),
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
			*evt.CorrelationEventID != "123" ||
			*evt.CausationEventID != "456" ||
			evt.Type != "SomeEvent" {

			t.Fatal("events not read")
		}
	}
}

func TestShouldWriteToDifferentStreamsWithTransaction(t *testing.T) {
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

	transactor := tx.New(gormtx.NewDB(es.DB))

	err := transactor.WithTransaction(ctx, func(ctx context.Context) error {
		err := es.AppendStream(
			ctx, streamOne, eventstore.InitialStreamVersion, toEventToStore(evts...),
		)
		if err != nil {
			t.Fatalf("error: %v", err)
		}

		err = es.AppendStream(
			ctx, streamTwo, eventstore.InitialStreamVersion, toEventToStore(evts...),
		)
		if err != nil {
			t.Fatalf("error: %v", err)
		}

		return nil
	})

	assert.NoError(t, err)
}

func TestShouldAppendToExistingStream(t *testing.T) {
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
		ctx, stream, eventstore.InitialStreamVersion, toEventToStore(evts...),
	)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	err = es.AppendStream(
		ctx, stream, 3, toEventToStore(evts...),
	)

	if err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestOptimisticConcurrencyCheckIsPerformed(t *testing.T) {
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
		ctx, stream, eventstore.InitialStreamVersion, toEventToStore(evts...),
	)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	err = es.AppendStream(
		ctx, stream, eventstore.InitialStreamVersion, toEventToStore(evts...),
	)

	if !errors.Is(err, eventstore.ErrConcurrencyCheckFailed) {
		t.Fatalf("should have performed optimistic concurrency check")
	}
}

func TestReadStreamWrapsNotFoundError(t *testing.T) {
	es, cleanup := eventStore(t)

	defer cleanup()

	_, err := es.ReadStream(context.Background(), "foo-stream")
	if !errors.Is(err, eventstore.ErrStreamNotFound) {
		t.Fatal("should return explicit error if stream doesn't exist")
	}
}

func TestSubscribeAllWithOffsetCatchesUpToNewEvents(t *testing.T) {
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

	err := es.AppendStream(ctx, "stream-one", eventstore.InitialStreamVersion, toEventToStore(evts...))
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

	err = es.AppendStream(ctx, "stream-two", eventstore.InitialStreamVersion, toEventToStore(evtsTwo...))
	if err != nil {
		t.Fatal(err)
	}

	got = readAllSub(t, sub, 4)

	if len(got) != 4 {
		t.Fatalf("should have read 4 events. actual: %d", len(got))
	}
}

func readAllSub(t *testing.T, sub eventstore.Subscription, expect int) []eventstore.StoredEvent {
	var got []eventstore.StoredEvent

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

	err := es.AppendStream(ctx, "stream-one", eventstore.InitialStreamVersion, toEventToStore(evts...))
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
		[]eventstore.EventToStore{
			{
				Event: SomeEvent{
					UserID: "123",
				},
			},
		},
	)

	if !errors.Is(err, anErr) {
		t.Fatal("error should have been propagated")
	}
}

func TestEncoderDecodeErrorsPropagated(t *testing.T) {
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
		[]eventstore.EventToStore{
			{
				Event: SomeEvent{
					UserID: "123",
				},
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
		[]eventstore.EventToStore{
			{
				Event: SomeEvent{
					UserID: "123",
				},
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
	_, err := eventstore.New(nil)
	if err == nil {
		t.Fatal("encoder must be provided")
	}
}

func TestNewDBMustBeProvided(t *testing.T) {
	_, err := eventstore.New(eventstore.NewJSONEncoder())
	if err == nil {
		t.Fatal("DB con must be provided")
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
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			err := es.AppendStream(context.Background(), tc.stream, tc.ver, toEventToStore(tc.evts))
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
	t.Helper()

	if *withPG {
		ctx := context.Background()

		dbName := "event-store"
		dbUser := "user"
		dbPassword := "password"

		postgresContainer, err := postgres.Run(
			ctx,
			"docker.io/postgres:16-alpine",
			postgres.WithDatabase(dbName),
			postgres.WithUsername(dbUser),
			postgres.WithPassword(dbPassword),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(5*time.Second)),
		)
		if err != nil {
			t.Fatal(err)
		}

		dsn, err := postgresContainer.ConnectionString(ctx)
		if err != nil {
			t.Fatal(err)
		}

		es, err := eventstore.New(enc, eventstore.WithPostgresDB(dsn))
		if err != nil {
			t.Fatalf("error creating es: %v", err)
		}

		return es, func() {
			if err := postgresContainer.Stop(ctx, nil); err != nil {
				// if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
				log.Fatalf("failed to terminate container: %s", err)
			}
		}
	}

	es, err := eventstore.New(enc, eventstore.WithSQLiteDB("file::memory:?cache=shared"))
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

func toEventToStore(events ...any) []eventstore.EventToStore {
	var evts []eventstore.EventToStore

	for _, evt := range events {
		evts = append(evts, eventstore.EventToStore{
			Event: evt,
			Meta: map[string]string{
				"ip": "127.0.0.1",
			},
			CorrelationEventID: "123",
			CausationEventID:   "456",
		})
	}

	return evts
}
