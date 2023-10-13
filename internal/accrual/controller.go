package accrual

import (
	"context"

	bonusesStorage "github.com/erupshis/bonusbridge/internal/bonuses/storage"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
	ordersStorage "github.com/erupshis/bonusbridge/internal/orders/storage"
)

type Controller struct {
	ordersStorage  ordersStorage.Storage
	bonusesStorage bonusesStorage.Storage

	log logger.BaseLogger
}

func CreateController(ordersStorage ordersStorage.Storage, bonusesStorage bonusesStorage.Storage, baseLogger logger.BaseLogger) Controller {
	return Controller{
		ordersStorage:  ordersStorage,
		bonusesStorage: bonusesStorage,
		log:            baseLogger,
	}
}

//TODO: need to run system. It should get queue of tasks to do and update if need.

func (c *Controller) Run(ctx context.Context) {
	ch := make(chan data.Order, 10)

	c.log.Info("[accrual:Controller:Run] start interaction with loyalty system")

	go c.requestCalculationsResult(ctx, ch)
	go c.updateOrders(ctx, ch)
}

func (c *Controller) requestCalculationsResult(ctx context.Context, chOut chan<- data.Order) {
	for {
		select {
		case <-ctx.Done():
			close(chOut)
			c.log.Info("[accrual:Controller:requestCalculationsResult] requests task is stopping")
			return
		default:
			ordersProcessing, err := c.ordersStorage.GetOrders(ctx, map[string]interface{}{"status_id": data.GetOrderStatusID("PROCESSING")})
			if err != nil {
				c.log.Info("[accrual:Controller:requestCalculationsResult] failed to get orders with PROCESSING status: %w", err)
			} else {
				for _, _ = range ordersProcessing {

				}
			}

			ordersNew, err := c.ordersStorage.GetOrders(ctx, map[string]interface{}{"status_id": data.GetOrderStatusID("NEW")})
			if err != nil {
				c.log.Info("[accrual:Controller:requestCalculationsResult] failed to get orders with NEW status: %w", err)
			} else {
				for _, _ = range ordersNew {

				}
			}
		}
	}
}

func (c *Controller) updateOrders(ctx context.Context, chOut <-chan data.Order) {
	for {
		select {
		case <-ctx.Done():
			c.log.Info("[accrual:Controller:updateOrders] update orders task is stopping")
			return
		default:
		}
	}
}
