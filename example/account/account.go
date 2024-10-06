package account

import (
	"github.com/aneshas/eventstore/aggregate"
)

// New creates new Account
func New(id, holder string) (*Account, error) {
	var acc Account

	acc.Rehydrate(&acc)

	acc.Apply(
		NewAccountOpened{
			ID:     id,
			Holder: holder,
		},
	)

	return &acc, nil
}

type ID string

// Account represents an account aggregate
type Account struct {
	aggregate.Root[ID]

	ID      string
	holder  string
	balance int
}

// Deposit money
func (a *Account) Deposit(amount int) {
	a.Apply(
		// TODO - Add ApplyMeta
		// internally Events() always should contain EventsToStore ?
		DepositMade{Amount: amount},
	)
}

// OnNewAccountOpened handler
// TODO - add second parameter
func (a *Account) OnNewAccountOpened(evt NewAccountOpened) {
	a.ID = evt.ID
	a.holder = evt.Holder
}

// OnDepositMade handler
func (a *Account) OnDepositMade(evt DepositMade) {
	a.balance += evt.Amount
}
