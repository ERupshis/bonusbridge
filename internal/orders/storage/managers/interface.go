package managers

import (
	"context"

	"github.com/erupshis/bonusbridge/internal/orders/data"
)

//go:generate mockgen -destination=../../../../mocks/mock_BaseOrdersManager.go -package=mocks github.com/erupshis/bonusbridge/internal/orders/storage/managers BaseOrdersManager
type BaseOrdersManager interface {
	AddOrder(ctx context.Context, number string, userID int64) (int64, error)
	UpdateOrder(ctx context.Context, order *data.Order) error
	GetOrders(ctx context.Context, filter map[string]interface{}) ([]data.Order, error)
}
