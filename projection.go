package eventstore

import (
	"context"
	"errors"
	"io"
	"log"
	"sync"
)

type EventStreamer interface {
	SubscribeAll(context.Context, ...SubAllOpt) (Subscription, error)
}

// eg. retry mechanism, error handler...
func NewProjector(s EventStreamer) *Projector {
	return &Projector{
		streamer: s,
		logger:   log.Default(),
	}
}

type Projector struct {
	streamer    EventStreamer
	projections []Projection
	logger      *log.Logger
}

type Projection func(EventData) error

func (p *Projector) Add(projections ...Projection) {
	p.projections = append(p.projections, projections...)
}

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

				if err := p.run(sub, projection); err != nil {
					continue
				}

				return
			}
		}(projection)
	}

	wg.Wait()

	return nil
}

func (p *Projector) run(sub Subscription, projection Projection) error {
	for {
		select {
		case data := <-sub.EventData:
			err := projection(data)
			if err != nil {
				p.logErr(err)
				// Retry
				// return
				// Retry then restart the whole projection after retry count is exceeded
				// (returning an err restarts the whole projection, returning nil exits)

				return err
			}

		case err := <-sub.Err:
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				// TODO - Should also exit if ctx is canceled
				if errors.Is(err, ErrSubscriptionClosedByClient) {
					return nil
				}

				p.logErr(err)
				// Retry
				// return
				// Retry then restart the whole projection after retry count is exceeded
				// (returning an err restarts the whole projection, returning nil exits)
			}
		}
	}
}

func (p *Projector) logErr(err error) {
	p.logger.Printf("projector error: %v", err)
}
