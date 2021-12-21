package eventstore_test

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/aneshas/eventstore"
)

type streamer struct {
	evts      []interface{}
	err       error
	streamErr error
	noClose   bool
	delay     *time.Duration
}

func (s streamer) SubscribeAll(ctx context.Context, opts ...eventstore.SubAllOpt) (eventstore.Subscription, error) {
	if s.err != nil {
		return eventstore.Subscription{}, s.err
	}

	sub := eventstore.Subscription{
		Err:       make(chan error, 1),
		EventData: make(chan eventstore.EventData),
	}

	go func() {
		if s.delay != nil {
			time.Sleep(*s.delay)
		}

		for _, evt := range s.evts {
			sub.EventData <- eventstore.EventData{
				Event: evt,
			}

			if s.streamErr != nil {
				sub.Err <- s.streamErr
				continue
			}

			sub.Err <- io.EOF
		}

		if !s.noClose {
			sub.Err <- eventstore.ErrSubscriptionClosedByClient
		}
	}()

	return sub, nil
}

func TestShouldProjectEventsToProjections(t *testing.T) {
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

	s := streamer{
		evts: evts,
	}

	p := eventstore.NewProjector(s)

	var got []interface{}
	var anotherGot []interface{}

	p.Add(
		func(ed eventstore.EventData) error {
			got = append(got, ed.Event)

			return nil
		},
		func(ed eventstore.EventData) error {
			anotherGot = append(anotherGot, ed.Event)

			return nil
		},
	)

	p.Run(context.TODO())

	if !reflect.DeepEqual(got, evts) ||
		!reflect.DeepEqual(anotherGot, evts) {
		t.Fatal("all projections should have received all events")
	}
}

func TestShouldRetryAndRestartIfProjectionErrorsOut(t *testing.T) {
	evts := []interface{}{
		SomeEvent{
			UserID: "user-1",
		},
	}

	s := streamer{
		evts: evts,
	}

	p := eventstore.NewProjector(s)

	var got []interface{}

	var times int

	p.Add(
		func(ed eventstore.EventData) error {
			if times < 3 {
				times++
				return fmt.Errorf("some transient error")
			}

			got = append(got, ed.Event)

			return nil
		},
	)

	p.Run(context.TODO())

	if !reflect.DeepEqual(got, evts) {
		t.Fatal("projection should have caught up after erroring out")
	}
}

func TestShouldRetrySubscriptionIfProjectionFailsToSubscribe(t *testing.T) {
	someErr := fmt.Errorf("some terminal error")

	s := streamer{
		err: someErr,
	}

	p := eventstore.NewProjector(s)

	p.Add(
		func(ed eventstore.EventData) error {
			return nil
		},
	)

	p.Run(context.TODO())
}

func TestShouldExitIfContextIsCanceled(t *testing.T) {
	evts := []interface{}{
		SomeEvent{
			UserID: "user-1",
		},
	}

	s := streamer{
		evts:    evts,
		noClose: true,
	}

	p := eventstore.NewProjector(s)

	p.Add(
		func(ed eventstore.EventData) error {
			return nil
		},
		func(ed eventstore.EventData) error {
			return nil
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)

	defer cancel()

	p.Run(ctx)
}

func TestShouldContinueProjectingIfStreamingErrorOccurs(t *testing.T) {
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

	s := streamer{
		evts:      evts,
		streamErr: fmt.Errorf("some error"),
	}

	p := eventstore.NewProjector(s)

	var got []interface{}

	p.Add(
		func(ed eventstore.EventData) error {
			got = append(got, ed.Event)

			return nil
		},
	)

	p.Run(context.TODO())

	if !reflect.DeepEqual(got, evts) {
		t.Fatal("projection should have caught up after erroring out")
	}
}

func TestShouldFlushProjection(t *testing.T) {
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

	d := 500 * time.Millisecond

	s := streamer{
		evts:  evts,
		delay: &d,
	}

	p := eventstore.NewProjector(s)

	var m sync.Mutex
	var got []interface{}

	called := false

	p.Add(
		eventstore.FlushAfter(
			func(ed eventstore.EventData) error {
				m.Lock()
				defer m.Unlock()

				got = append(got, ed.Event)

				return nil
			},
			func() error {
				m.Lock()
				defer m.Unlock()

				called = true

				return nil
			},
			200*time.Millisecond,
		),
	)

	p.Run(context.TODO())

	<-time.After(1000 * time.Millisecond)

	m.Lock()
	defer m.Unlock()

	if !reflect.DeepEqual(got, evts) {
		t.Fatal("projection should have received all events")
	}

	if !called {
		t.Fatal("flush should have been called")
	}
}
