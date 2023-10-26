package data

import (
	"fmt"
	"time"
)

var ErrOrderWasAddedByAnotherUser = fmt.Errorf("order has already been added by another user")
var ErrOrderWasAddedBefore = fmt.Errorf("order has already been added")

const (
	StatusNew = iota + 1
	StatusProcessing
	StatusInvalid
	StatusProcessed
	StatusUndefined
)

func GetOrderStatusID(statusStr string) int {
	res := StatusUndefined
	switch statusStr {
	case "NEW":
		res = StatusNew
	case "PROCESSING":
		res = StatusProcessing
	case "INVALID":
		res = StatusInvalid
	case "PROCESSED":
		res = StatusProcessed
	default:
		res = StatusUndefined
	}

	return res
}

//go:generate easyjson -all data.go
type Order struct {
	ID         int       `json:"-"`
	Number     string    `json:"number"`
	UserID     int64     `json:"-"`
	BonusID    int64     `json:"-"`
	Status     string    `json:"status"`
	Accrual    float32   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
