package ambar

import (
	"context"
	"encoding/json"
	"github.com/aneshas/eventstore"
)

// TODO - Define different error types eg - retry - for ambar policies

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

type req struct {
	Payload payload `json:"payload"`
}

type payload struct {
	Event              string  `json:"data"`
	Meta               string  `json:"meta"`
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
func (a *Ambar) Project(_ context.Context, projection eventstore.Projection, data []byte) error {
	var event req

	err := json.Unmarshal(data, &event)
	if err != nil {
		// retry or ?
		return err
	}

	decoded, err := a.dec.Decode(&eventstore.EncodedEvt{
		Data: event.Payload.Event,
		Type: event.Payload.Type,
	})
	if err != nil {
		// if error is no event registered - ignore
		return err
	}

	// parse occurred_on
	// parse meta

	return projection(eventstore.StoredEvent{
		Event:              decoded,
		ID:                 event.Payload.ID,
		Sequence:           event.Payload.Sequence,
		Type:               event.Payload.Type,
		CausationEventID:   event.Payload.CausationEventID,
		CorrelationEventID: event.Payload.CorrelationEventID,
		StreamID:           event.Payload.StreamID,
		StreamVersion:      event.Payload.StreamVersion,
	})
}
