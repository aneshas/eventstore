package account

// Events is a list of all domain event instances
var Events = []any{
	NewAccountOpened{},
	DepositMade{},
}

// NewAccountOpened domain event indicates that new
// account has been opened
type NewAccountOpened struct {
	AccountID string
	Holder    string
}

// DepositMade domain event indicates that deposit has been made
type DepositMade struct {
	Amount int
}
