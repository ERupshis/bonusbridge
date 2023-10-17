package data

import (
	"fmt"
	"time"
)

var ErrNotEnoughBonuses = fmt.Errorf("not enough bonuses for withdrawal")
var ErrWithdrawalsMissing = fmt.Errorf("user doesn't have any withdrawal")

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
	BonusID     int64     `json:"-"`
	Order       string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
