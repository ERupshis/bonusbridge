package data

import (
	"time"
)

//go:generate easyjson -all data.go
type Order struct {
	ID         int       `json:"-"`
	Number     string    `json:"number"`
	UserID     int64     `json:"-"`
	Status     string    `json:"status"`
	Accrual    string    `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
