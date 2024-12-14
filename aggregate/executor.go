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
