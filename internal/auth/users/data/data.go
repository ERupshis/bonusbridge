package data

import (
	"fmt"
)

const UserID = "userID"

const (
	RoleUser = iota + 1
	RoleAdmin
)

// ErrUserNotFound missing user in database.
var ErrUserNotFound = fmt.Errorf("user not found")

// User represents a user in our system.
//
//go:generate easyjson -all data.go
type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`

	ID   int64 `json:"-"`
	Role int   `json:"-"`
}
