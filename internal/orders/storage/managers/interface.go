package managers

import (
	"context"

	"github.com/erupshis/bonusbridge/internal/orders/data"
)

type BaseOrdersManager interface {
	AddOrder(ctx context.Context, number string, userID int64) (int64, error)
	GetOrder(ctx context.Context, number string) (*data.Order, error)
	GetOrders(ctx context.Context, userID int64) ([]data.Order, error)
}
