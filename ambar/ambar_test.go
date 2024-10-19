package ambar_test

import (
	"encoding/json"
	"github.com/aneshas/eventstore"
	"github.com/aneshas/eventstore/ambar"
	"github.com/aneshas/eventstore/ambar/testutil"
	"github.com/relvacode/iso8601"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShould_Project_Required_Data(t *testing.T) {
	var a = ambar.New(eventstore.NewJSONEncoder(testutil.TestEvent{}))

	occurredOn, err := iso8601.ParseString(testutil.AmbarPayload.OccurredOn)
	if err != nil {
		t.Fatal(err)
	}

	projection := func(data eventstore.StoredEvent) error {
		assert.Equal(t, eventstore.StoredEvent{
			Event:              testutil.Event,
			Meta:               nil,
			ID:                 testutil.AmbarPayload.ID,
			Sequence:           testutil.AmbarPayload.Sequence,
			Type:               "TestEvent",
			CausationEventID:   nil,
			CorrelationEventID: nil,
			StreamID:           testutil.AmbarPayload.StreamID,
			StreamVersion:      testutil.AmbarPayload.StreamVersion,
			OccurredOn:         occurredOn,
		}, data)

		return nil
	}

	err = a.Project(nil, projection, testutil.Payload(t, testutil.AmbarPayload))

	assert.NoError(t, err)
}

func TestShould_Retry_On_Bad_Date_Format(t *testing.T) {
	p := testutil.AmbarPayload

	p.OccurredOn = "bad-date-time"

	var a = ambar.New(eventstore.NewJSONEncoder(testutil.TestEvent{}))

	err := a.Project(nil, nil, testutil.Payload(t, p))

	assert.Error(t, err)
}

func TestShould_Retry_On_Bad_Meta_Format(t *testing.T) {
	p := testutil.AmbarPayload

	badMeta := "bad-meta"

	p.Meta = &badMeta

	var a = ambar.New(eventstore.NewJSONEncoder(testutil.TestEvent{}))

	err := a.Project(nil, nil, testutil.Payload(t, p))

	assert.Error(t, err)
}

func TestShould_Not_Retry_On_Unregistered_Event(t *testing.T) {
	var a = ambar.New(eventstore.NewJSONEncoder())

	err := a.Project(nil, nil, testutil.Payload(t, testutil.AmbarPayload))

	assert.NoError(t, err)
}

func TestShould_Project_Optional_Data(t *testing.T) {
	p := testutil.AmbarPayload

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

	var a = ambar.New(eventstore.NewJSONEncoder(testutil.TestEvent{}))

	projection := func(data eventstore.StoredEvent) error {
		assert.Equal(t, correlationEventID, *data.CorrelationEventID)
		assert.Equal(t, causationEventID, *data.CausationEventID)
		assert.Equal(t, meta, data.Meta)

		return nil
	}

	err = a.Project(nil, projection, testutil.Payload(t, p))

	assert.NoError(t, err)
}
