package ambar

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aneshas/eventstore"
	"github.com/relvacode/iso8601"
)

var (
	// ErrNoRetry is the error returned when we don't want to retry
	// projecting events in case of an error.
	// This is also the default behavior when an error is returned but this
	// error can be used if we also want to wrap the error eg. for logging
	ErrNoRetry = errors.New("retry")

	// ErrKeepItGoing is the error returned when we want to keep projecting
	// events in case of an error
	ErrKeepItGoing = errors.New("keep it going")
)

// SuccessResp is the success response
// https://docs.ambar.cloud/#Data%20Destinations
var SuccessResp = `{
  "result": {
    "success": {}
  }
}`

// RetryResp is the retry response
// https://docs.ambar.cloud/#Data%20Destinations
var RetryResp = `{
  "result": {
    "error": {
      "policy": "must_retry", 
      "class": "must retry it", 
      "description": "must retry it"
    }
  }
}`

// KeepGoingResp is the keep going response
// https://docs.ambar.cloud/#Data%20Destinations
var KeepGoingResp = `{
  "result": {
    "error": {
      "policy": "keep_going", 
      "class": "keep it going", 
      "description": "keep it going"
    }
  }
}`

// New constructs a new Ambar projection handler
func New(dec Decoder) *Ambar {
	return &Ambar{dec: dec}
}

// Decoder is an interface for decoding events
type Decoder interface {
	Decode(*eventstore.EncodedEvt) (any, error)
}

// Ambar is a projection handler for ambar events
type Ambar struct {
	dec Decoder
}

// Req is the ambar projection request
type Req struct {
	Payload Payload `json:"payload"`
}

// Payload is the ambar projection request payload
type Payload struct {
	Event              string  `json:"data"`
	Meta               *string `json:"meta"`
	ID                 string  `json:"id"`
	Sequence           uint64  `json:"sequence"`
	Type               string  `json:"type"`
	CausationEventID   *string `json:"causation_event_id"`
	CorrelationEventID *string `json:"correlation_event_id"`
	StreamID           string  `json:"stream_id"`
	StreamVersion      int     `json:"stream_version"`
	OccurredOn         string  `json:"occurred_on"`
}

// Project projects ambar event to provided projection
// It will always return ambar retry policy error if deserialization fails
func (a *Ambar) Project(_ context.Context, projection eventstore.Projection, data []byte) error {
	var event Req

	err := json.Unmarshal(data, &event)
	if err != nil {
		return err
	}

	decoded, err := a.dec.Decode(&eventstore.EncodedEvt{
		Data: event.Payload.Event,
		Type: event.Payload.Type,
	})
	if err != nil {
		if errors.Is(err, eventstore.ErrEventNotRegistered) {
			return nil
		}

		return err
	}

	occurredOn, err := iso8601.ParseString(event.Payload.OccurredOn)
	if err != nil {
		return err
	}

	var meta map[string]string

	if event.Payload.Meta != nil {
		err = json.Unmarshal([]byte(*event.Payload.Meta), &meta)
		if err != nil {
			return err
		}
	}

	return projection(eventstore.StoredEvent{
		Event:              decoded,
		ID:                 event.Payload.ID,
		Meta:               meta,
		Sequence:           event.Payload.Sequence,
		Type:               event.Payload.Type,
		CausationEventID:   event.Payload.CausationEventID,
		CorrelationEventID: event.Payload.CorrelationEventID,
		StreamID:           event.Payload.StreamID,
		StreamVersion:      event.Payload.StreamVersion,
		OccurredOn:         occurredOn,
	})
}
