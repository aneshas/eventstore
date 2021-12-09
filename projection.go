package eventstore

import (
	"context"
	"sync"
)

// TODO Add common middlewares and ability to wrap
// eg. retry mechanism, error handler...
type Projection func(EventData) error

// TODO configure retries and logging
// and readAll options

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
			// TODO - Offset tracking (later)
			sub, _ := p.store.ReadAll(ctx)

			defer sub.Close()

			p.run(sub, projection)
		}(projection)
	}

	wg.Wait()

	// TODO - Collect errors

	return nil
}

func (p *Projector) run(sub Subscription, projection Projection) {
	for {
		select {
		case data := <-sub.EventData:
			err := projection(data)
			if err != nil {
				// TODO - Handle retries, logging etc...
			}

		case err := <-sub.Err:
			// Check if error is io.EOF and decide weather to continue
			// consider throttling after EOF

			if err != nil {
				// If eof handle retry / logging
			}
		}
	}
}

// TODO - Design projetions to work in batches like kafka eg.
// we would need begin and commit hooks
