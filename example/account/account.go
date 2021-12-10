package account

import "github.com/aneshas/eventstore"

// New creates new Account
func New(id, holder string) (*Account, error) {
	var acc Account

	err := acc.Init(&acc)
	if err != nil {
		return nil, err
	}

	err = acc.Apply(
		NewAccountOpenned{
			ID:     id,
			Holder: holder,
		},
	)
	if err != nil {
		return nil, err
	}

	return &acc, err
}

// NewAccountOpenned domain event indicates that new
// account has been openned
type NewAccountOpenned struct {
	ID     string
	Holder string
}

// Account represents an account domain model
type Account struct {
	eventstore.AggregateRoot

	ID     string
	holder string
}

// OnNewAccountOpenned event handler
func (a *Account) OnNewAccountOpenned(evt NewAccountOpenned) error {
	a.ID = evt.ID
	a.holder = evt.Holder

	return nil
}
