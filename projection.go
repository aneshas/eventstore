package eventstore

import (
	"context"
	"errors"
	"io"
	"log"
	"sync"
)

// TODO Add common middlewares and ability to wrap
// eg. retry mechanism, error handler...
type Projection func(EventData) error

// and readAll options

// Rename to SimpleProjector ?
func NewProjector(store *EventStore) *Projector {
	return nil
}

type Projector struct {
	store       *EventStore
	projections []Projection
}

func (p *Projector) Add(projections ...Projection) {
	p.projections = append(p.projections, projections...)
}

func (p *Projector) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	for _, projection := range p.projections {
		wg.Add(1)

		go func(projection Projection) {
			sub, _ := p.store.ReadAll(ctx)

			defer sub.Close()

			p.run(sub, projection)
		}(projection)
	}

	wg.Wait()

	return nil
}

func (p *Projector) run(sub Subscription, projection Projection) {
	for {
		select {
		case data := <-sub.EventData:
			// TODO - Loop while err != nil
			err := projection(data)
			if err != nil {
				// Log
				log.Println(err)
			}

		case err := <-sub.Err:
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				log.Println(err)
			}
		}
	}
}
