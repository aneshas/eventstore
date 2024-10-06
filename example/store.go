package example

import (
	"context"

	"github.com/aneshas/eventstore"
	"github.com/aneshas/eventstore-example/account"
)

// NewAccountStore constructs new account repository
func NewAccountStore(eventStore *eventstore.EventStore) *AccountStore {
	return &AccountStore{
		eventStore: eventStore,
	}
}

// AccountStore represents account eventstore repository implementation
// which you might implement in order to abstract the underlying persistent mechanism
// In essence all event sourcing is is a different form of persistence
type AccountStore struct {
	eventStore *eventstore.EventStore
}

// Save saves an account
func (s *AccountStore) Save(ctx context.Context, acc *account.Account) error {
	var events []eventstore.EventToStore

	for _, evt := range acc.Events() {
		events = append(events, eventstore.EventToStore{
			Event: evt,

			// Optional
			ID:                 "",
			CausationEventID:   "",
			CorrelationEventID: "",
			Meta:               nil,
		})
	}

	return s.eventStore.AppendStream(
		ctx,
		acc.ID,
		acc.Version(),
		events,
	)
}

// FindByID finds an account by it's id
func (s *AccountStore) FindByID(ctx context.Context, id string) (*account.Account, error) {
	storedEvents, err := s.eventStore.ReadStream(ctx, id)
	if err != nil {
		return nil, err
	}

	var events []any

	for _, evt := range storedEvents {
		// eg. extract additional props (CreatedAt etc...)
		events = append(events, evt.Event)
	}

	var acc account.Account

	// TODO - Should accept stored event and call alternate apply method
	acc.Rehydrate(&acc, events...)

	return &acc, nil
}
