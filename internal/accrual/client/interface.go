package client

import (
	"context"

	"github.com/erupshis/bonusbridge/internal/orders/data"
)

type ResponseStatus int
type RetryInterval int

type BaseClient interface {
	RequestCalculationResult(ctx context.Context, host string, order *data.Order) (ResponseStatus, RetryInterval, error)
}
