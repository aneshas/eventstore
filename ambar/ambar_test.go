package ambar_test

import (
	"encoding/json"
	"github.com/aneshas/eventstore"
	"github.com/aneshas/eventstore/ambar"
	"github.com/relvacode/iso8601"
	"github.com/stretchr/testify/assert"
	"testing"
)

type anEvent struct {
	Foo string
	Bar string
}

var event = anEvent{
	Foo: "foo",
	Bar: "bar",
}

var ambarPayload = ambar.Payload{
	Event:              eventData(),
	Meta:               nil,
	ID:                 "event-id",
	Sequence:           1,
	Type:               "anEvent",
	CausationEventID:   nil,
	CorrelationEventID: nil,
	StreamID:           "stream-id",
	StreamVersion:      1,
	OccurredOn:         "2024-10-12T20:07:22.436271+00",
}

func TestShould_Project_Required_Data(t *testing.T) {
	var a = ambar.New(eventstore.NewJSONEncoder(anEvent{}))

	occurredOn, err := iso8601.ParseString(ambarPayload.OccurredOn)
	if err != nil {
		t.Fatal(err)
	}

	projection := func(data eventstore.StoredEvent) error {
		assert.Equal(t, eventstore.StoredEvent{
			Event:              event,
			Meta:               nil,
			ID:                 ambarPayload.ID,
			Sequence:           ambarPayload.Sequence,
			Type:               "anEvent",
			CausationEventID:   nil,
			CorrelationEventID: nil,
			StreamID:           ambarPayload.StreamID,
			StreamVersion:      ambarPayload.StreamVersion,
			OccurredOn:         occurredOn,
		}, data)

		return nil
	}

	err = a.Project(nil, projection, payload(t, ambarPayload))

	assert.NoError(t, err)
}

func TestShould_Retry_On_Bad_Date_Format(t *testing.T) {
	p := ambarPayload

	p.OccurredOn = "bad-date-time"

	var a = ambar.New(eventstore.NewJSONEncoder(anEvent{}))

	err := a.Project(nil, nil, payload(t, p))

	assert.ErrorIs(t, err, ambar.ErrRetry)
}

func TestShould_Retry_On_Bad_Meta_Format(t *testing.T) {
	p := ambarPayload

	badMeta := "bad-meta"

	p.Meta = &badMeta

	var a = ambar.New(eventstore.NewJSONEncoder(anEvent{}))

	err := a.Project(nil, nil, payload(t, p))

	assert.ErrorIs(t, err, ambar.ErrRetry)
}

func TestShould_Not_Retry_On_Unregistered_Event(t *testing.T) {
	var a = ambar.New(eventstore.NewJSONEncoder())

	err := a.Project(nil, nil, payload(t, ambarPayload))

	assert.NoError(t, err)
}

func TestShould_Project_Optional_Data(t *testing.T) {
	p := ambarPayload

	meta := map[string]string{
		"foo": "bar",
	}

	metaData, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}

	metaStr := string(metaData)
	correlationEventID := "correlation-event-id"
	causationEventID := "causation-event-id"

	p.Meta = &metaStr
	p.CausationEventID = &causationEventID
	p.CorrelationEventID = &correlationEventID

	var a = ambar.New(eventstore.NewJSONEncoder(anEvent{}))

	projection := func(data eventstore.StoredEvent) error {
		assert.Equal(t, correlationEventID, *data.CorrelationEventID)
		assert.Equal(t, causationEventID, *data.CausationEventID)
		assert.Equal(t, meta, data.Meta)

		return nil
	}

	err = a.Project(nil, projection, payload(t, p))

	assert.NoError(t, err)
}

func payload(t *testing.T, p ambar.Payload) []byte {
	t.Helper()

	data, err := json.Marshal(ambar.Req{
		Payload: p,
	})
	if err != nil {
		t.Fatal(err)
	}

	return data
}

func eventData() string {
	data, err := json.Marshal(event)
	if err != nil {
		panic(err)
	}

	return string(data)
}
