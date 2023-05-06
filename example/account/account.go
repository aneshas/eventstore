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
		NewAccountOpened{
			ID:     id,
			Holder: holder,
		},
	)
	if err != nil {
		return nil, err
	}

	return &acc, err
}

// NewAccountOpened domain event indicates that new
// account has been opened
type NewAccountOpened struct {
	ID     string
	Holder string
}

// Account represents an account domain model
type Account struct {
	eventstore.AggregateRoot

	ID     string
	holder string
}

// OnNewAccountOpened event handler
func (a *Account) OnNewAccountOpened(evt NewAccountOpened) {
	a.ID = evt.ID
	a.holder = evt.Holder
}
