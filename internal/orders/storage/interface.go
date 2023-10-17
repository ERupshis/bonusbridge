package storage

import (
	"context"

	"github.com/erupshis/bonusbridge/internal/orders/data"
)

type BaseOrdersStorage interface {
	AddOrder(ctx context.Context, number string, userID int64) error
	UpdateOrder(ctx context.Context, order *data.Order) error
	GetOrders(ctx context.Context, filter map[string]interface{}) ([]data.Order, error)
}
