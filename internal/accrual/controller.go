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
			//TODO: need to return await time and wait it to continue.
			for _, status := range []string{"PROCESSING", "NEW"} {
				orders, err := c.ordersStorage.GetOrders(ctx, map[string]interface{}{"status_id": data.GetOrderStatusID(status)})
				if err != nil {
					c.log.Info("[accrual:Controller:requestCalculationsResult] failed to get orders with PROCESSING status: %w", err)
				} else {
					for i := 0; i < len(orders); i++ {

					}
				}
			}
		}
	}
}

func (c *Controller) updateOrders(ctx context.Context, chIn <-chan data.Order) {
	for {
		select {
		case <-ctx.Done():
			c.log.Info("[accrual:Controller:updateOrders] update orders task is stopping")
			return
		case order := <-chIn:
			if data.GetOrderStatusID(order.Status) > data.StatusProcessing {
				if err := c.ordersStorage.UpdateOrder(ctx, &order); err != nil {
					c.log.Info("[accrual:Controller:updateOrders] error occurred during order '%v' update in db: %w", order, err)
				}
			}
		}
	}
}
