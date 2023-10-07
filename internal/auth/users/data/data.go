package data

import (
	"fmt"
)

// ErrUserNotFound missing user in database.
var ErrUserNotFound = fmt.Errorf("user not found")

// User represents a user in our system.
type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`

	ID   int
	Role int
}
