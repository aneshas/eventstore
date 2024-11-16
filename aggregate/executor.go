package aggregate

import (
	"context"
)

// NewExecutor creates a new executor for the given aggregate store.
func NewExecutor[T Rooter](store *Store[T]) Executor[T] {
	return func(ctx context.Context, a T, f func(ctx context.Context) error) error {
		return Exec(ctx, store, a, f)
	}
}

// Executor is a helper function to load an aggregate from the store, execute a function and save the aggregate back to the store.
type Executor[T Rooter] func(ctx context.Context, a T, f func(ctx context.Context) error) error

// Exec is a helper function to load an aggregate from the store, execute a function and save the aggregate back to the store.
func Exec[T Rooter](ctx context.Context, store *Store[T], a T, f func(ctx context.Context) error) error {
	err := store.ByID(ctx, a.StringID(), a)
	if err != nil {
		return err
	}

	err = f(ctx)
	if err != nil {
		return err
	}

	return store.Save(ctx, a)
}

// TODO
// always start a transaction in s.AppendEvents if one not present in the context - if yes then use it
// this way we can ensure that all events are saved in a single transaction
// and in the case of multiple aggregates we can ensure that all aggregates are saved in a single transaction
// also then client need to wrap the call to s.AppendEvents in a transaction only if they want to save multiple aggregates in a single transaction
