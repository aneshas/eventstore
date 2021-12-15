package eventstore

import (
	"context"
	"errors"
	"io"
	"log"
	"sync"
)

// EventStreamer represents an event stream that can be subscribed to
// This package offers EventStore as EventStreamer implementation
type EventStreamer interface {
	SubscribeAll(context.Context, ...SubAllOpt) (Subscription, error)
}

// NewProjector constructs a Projector
// TODO Configure logger, and retry
func NewProjector(s EventStreamer) *Projector {
	return &Projector{
		streamer: s,
		logger:   log.Default(),
	}
}

// Projector is an event projector which will subscribe to an
// event stream (evet store) and project events to each
// individual projection in an asynchronous manner
type Projector struct {
	streamer    EventStreamer
	projections []Projection
	logger      *log.Logger
}

// Projection represents a projection that should be able to handle
// projected events
type Projection func(EventData) error

// Add effectively registers a projection with the projector
// Make sure to add all of your projections before calling Run
func (p *Projector) Add(projections ...Projection) {
	p.projections = append(p.projections, projections...)
}

// Run will start the projector
func (p *Projector) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	for _, projection := range p.projections {
		wg.Add(1)

		go func(projection Projection) {
			defer wg.Done()

			for {
				// TODO retry with backoff
				sub, err := p.streamer.SubscribeAll(ctx)
				if err != nil {
					p.logErr(err)

					return
				}

				defer sub.Close()

				if err := p.run(ctx, sub, projection); err != nil {
					continue
				}

				return
			}
		}(projection)
	}

	wg.Wait()

	return nil
}

func (p *Projector) run(ctx context.Context, sub Subscription, projection Projection) error {
	for {
		select {
		case data := <-sub.EventData:
			err := projection(data)
			if err != nil {
				p.logErr(err)
				// TODO retry with backoff

				return err
			}

		case err := <-sub.Err:
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				if errors.Is(err, ErrSubscriptionClosedByClient) {
					return nil
				}

				p.logErr(err)
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (p *Projector) logErr(err error) {
	p.logger.Printf("projector error: %v", err)
}
