package account

import (
	"fmt"
	"github.com/aneshas/eventstore/aggregate"
)

// New opens a new Account
func New(id ID, holder string) (*Account, error) {
	var acc Account

	// We always need to call Rehydrate on a fresh instance in order to initialize the aggregate
	// so the events can be applied to it properly
	acc.Rehydrate(&acc)

	acc.Apply(
		NewAccountOpened{
			AccountID: id.String(),
			Holder:    holder,
		},
	)

	return &acc, nil
}

// Account represents an account aggregate
type Account struct {
	aggregate.Root[ID]

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

// Withdraw money
func (a *Account) Withdraw(amount int) error {
	if a.Balance < amount {
		return fmt.Errorf("insufficient funds")
	}

	a.Apply(
		WithdrawalMade{
			Amount: amount,
		},
	)

	return nil
}

// OnNewAccountOpened handler
func (a *Account) OnNewAccountOpened(evt NewAccountOpened) {
	a.SetID(ParseID(evt.AccountID))
}

// OnDepositMade handler
func (a *Account) OnDepositMade(evt DepositMade) {
	a.Balance += evt.Amount
}

// OnWithdrawalMade handler
func (a *Account) OnWithdrawalMade(evt WithdrawalMade) {
	a.Balance -= evt.Amount
}
