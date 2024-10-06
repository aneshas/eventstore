package account

// NewAccountOpened domain event indicates that new
// account has been opened
type NewAccountOpened struct {
	ID     string
	Holder string
}

// DepositMade domain event indicates that deposit has been made
type DepositMade struct {
	Amount int
}
