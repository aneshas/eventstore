package testutil

import (
	"encoding/json"
	"github.com/aneshas/eventstore/ambar"
	"testing"
)

// TestEvent is a test event
type TestEvent struct {
	Foo string
	Bar string
}

// Event is an instance of a test event
var Event = TestEvent{
	Foo: "foo",
	Bar: "bar",
}

// AmbarPayload is a test payload
var AmbarPayload = ambar.Payload{
	Event:              eventData(),
	Meta:               nil,
	ID:                 "event-id",
	Sequence:           1,
	Type:               "TestEvent",
	CausationEventID:   nil,
	CorrelationEventID: nil,
	StreamID:           "stream-id",
	StreamVersion:      1,
	OccurredOn:         "2024-10-12T20:07:22.436271+00",
}

func eventData() string {
	data, err := json.Marshal(Event)
	if err != nil {
		panic(err)
	}

	return string(data)
}

// Payload creates a payload for testing
func Payload(t *testing.T, p ambar.Payload) []byte {
	t.Helper()

	data, err := json.Marshal(ambar.Req{
		Payload: p,
	})
	if err != nil {
		t.Fatal(err)
	}

	return data
}
