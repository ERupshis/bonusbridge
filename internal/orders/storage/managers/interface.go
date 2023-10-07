package managers

import (
	"github.com/erupshis/bonusbridge/internal/orders/storage/data"
)

type BaseStorageManager interface {
	AddOrder(number string, userID int64) error
	GetOrders(userID int64) ([]data.Order, error)
}
