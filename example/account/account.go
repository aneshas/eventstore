package account

import (
	"github.com/aneshas/eventstore/aggregate"
)

// New creates new Account
func New(id ID, holder string) (*Account, error) {
	var acc Account

	acc.Rehydrate(&acc)

	acc.Apply(
		NewAccountOpened{
			AccountID: id.String(),
			Holder:    holder,
		},
	)

	return &acc, nil
}

// ID represents an account ID
type ID string

// String implements fmt.Stringer
func (id ID) String() string { return string(id) }

// Account represents an account aggregate
type Account struct {
	aggregate.Root[ID]

	// notice how aggregate has no state until it is needed to make a decision

	Balance int
}

// Deposit money
func (a *Account) Deposit(amount int) {
	// for example: check if amount is positive or account is not closed

	// if all good, do the mutation by applying the event
	a.Apply(
		DepositMade{
			Amount: amount,
		},
	)
}

// OnNewAccountOpened handler
func (a *Account) OnNewAccountOpened(evt NewAccountOpened) {
	a.SetID(ID(evt.AccountID))
}

// OnDepositMade handler
func (a *Account) OnDepositMade(evt DepositMade) {
	a.Balance += evt.Amount
}
