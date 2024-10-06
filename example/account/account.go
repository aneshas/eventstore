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

	holder  string
	balance int
}

// Deposit money
func (a *Account) Deposit(amount int) {
	a.Apply(
		// TODO - From aggregate we should always set the event ID
		// and occurred on ?
		// or
		//
		// Apply always sets ID and occured on internally
		// Provide alternate Apply with IDs, OccuredOn - make this one accept single event (bcs of id)
		// and others (correlation, meta, pass through context)

		DepositMade{
			Amount: amount,
		},
	)
}

// OnNewAccountOpened handler
func (a *Account) OnNewAccountOpened(evt NewAccountOpened) {
	a.SetID(ID(evt.AccountID))
	a.holder = evt.Holder
}

// OnDepositMade handler
func (a *Account) OnDepositMade(evt DepositMade) {
	a.balance += evt.Amount
}
