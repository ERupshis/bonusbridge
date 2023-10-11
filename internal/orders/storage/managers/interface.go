package managers

import (
	"context"

	"github.com/erupshis/bonusbridge/internal/orders/data"
)

type BaseStorageManager interface {
	AddOrder(ctx context.Context, number string, userID int64) (int64, error)
	GetOrder(ctx context.Context, number string) (*data.Order, error)
	GetOrders(userID int64) ([]data.Order, error)
}
