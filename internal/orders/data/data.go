package data

import (
	"time"
)

const (
	statusNew = iota + 1
	statusProcessing
	statusInvalid
	statusProcessed
	statusUndefined
)

func GetOrderStatusID(statusStr string) int {
	res := statusUndefined
	switch statusStr {
	case "NEW":
		res = statusNew
	case "PROCESSING":
		res = statusProcessing
	case "INVALID":
		res = statusInvalid
	case "PROCESSED":
		res = statusProcessed
	default:
		res = statusUndefined
	}

	return res
}

//go:generate easyjson -all data.go
type Order struct {
	ID         int       `json:"-"`
	Number     string    `json:"number"`
	UserID     int64     `json:"-"`
	Status     string    `json:"status"`
	Accrual    int       `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
