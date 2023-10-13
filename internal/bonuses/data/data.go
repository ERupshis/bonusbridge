package data

import (
	"time"
)

//go:generate easyjson -all data.go
type Balance struct {
	ID        int64   `json:"-"`
	UserID    int64   `json:"-"`
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type Withdrawal struct {
	ID          int64     `json:"-"`
	UserID      int64     `json:"-"`
	Order       string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
