package account

import "github.com/google/uuid"

// NewID generates a new account ID
func NewID() ID {
	return ID{uuid.New()}
}

// ID represents an account ID
type ID struct {
	uuid.UUID
}

// ParseID parses account ID from string
func ParseID(id string) ID {
	return ID{uuid.MustParse(id)}
}
