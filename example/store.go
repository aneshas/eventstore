package example

import (
	"context"

	"github.com/aneshas/eventstore"
	"github.com/aneshas/eventstore/example/account"
)

// NewAccountStore constructs new account repository
func NewAccountStore(estore *eventstore.EventStore) *AccountStore {
	return &AccountStore{
		estore: estore,
	}
}

// AccountStore represents account eventstore repository implementation
// which you might implement in order to abstract the underlying persistent mechanism
// In essence all event sourcing is is a different form of persistence
type AccountStore struct {
	estore *eventstore.EventStore
}

// Save saves an account
func (s *AccountStore) Save(ctx context.Context, acc *account.Account) error {
	return s.estore.AppendStream(
		ctx,
		acc.ID,
		acc.Version(),
		acc.Events(),
		// eg. read meta data from ctx
		// such as ip, username etc...
		eventstore.WithMetaData(nil),
	)
}

// FindByID finds an account by it's id
func (s *AccountStore) FindByID(ctx context.Context, id string) (*account.Account, error) {
	data, err := s.estore.ReadStream(ctx, id)
	if err != nil {
		return nil, err
	}

	var evts []interface{}

	for _, evt := range data {
		// eg. extract additional props (CreatedAt etc...)
		evts = append(evts, evt.Event)
	}

	var acc account.Account

	return &acc, acc.Init(&acc, evts...)
}
