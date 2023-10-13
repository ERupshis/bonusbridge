package client

import (
	"context"

	"github.com/erupshis/bonusbridge/internal/orders/data"
)

type BaseClient interface {
	RequestCalculationResult(ctx context.Context, order *data.Order) (int, error)
}
