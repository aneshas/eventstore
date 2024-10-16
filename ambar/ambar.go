package ambar

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aneshas/eventstore"
	"github.com/relvacode/iso8601"
)

// ErrRetry is the error returned when a retry is required
var ErrRetry = errors.New("retry")

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
// It will always return ambar retry policy error if deserilization fails
func (a *Ambar) Project(_ context.Context, projection eventstore.Projection, data []byte) error {
	var event Req

	err := json.Unmarshal(data, &event)
	if err != nil {
		return errors.Join(err, ErrRetry)
	}

	decoded, err := a.dec.Decode(&eventstore.EncodedEvt{
		Data: event.Payload.Event,
		Type: event.Payload.Type,
	})
	if err != nil {
		if errors.Is(err, eventstore.ErrEventNotRegistered) {
			return nil
		}

		return errors.Join(err, ErrRetry)
	}

	occurredOn, err := iso8601.ParseString(event.Payload.OccurredOn)
	if err != nil {
		return errors.Join(err, ErrRetry)
	}

	var meta map[string]string

	if event.Payload.Meta != nil {
		err = json.Unmarshal([]byte(*event.Payload.Meta), &meta)
		if err != nil {
			return errors.Join(err, ErrRetry)
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
