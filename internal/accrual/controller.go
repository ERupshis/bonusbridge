package accrual

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/erupshis/bonusbridge/internal/accrual/client"
	bonusesStorage "github.com/erupshis/bonusbridge/internal/bonuses/storage"
	"github.com/erupshis/bonusbridge/internal/config"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
	ordersStorage "github.com/erupshis/bonusbridge/internal/orders/storage"
)

type Controller struct {
	ordersStorage  ordersStorage.Storage
	bonusesStorage bonusesStorage.Storage

	client client.BaseClient

	accrualAddr string

	log logger.BaseLogger
}

func CreateController(ordersStorage ordersStorage.Storage, bonusesStorage bonusesStorage.Storage, client client.BaseClient, cfg config.Config, baseLogger logger.BaseLogger) Controller {
	return Controller{
		ordersStorage:  ordersStorage,
		bonusesStorage: bonusesStorage,
		client:         client,
		accrualAddr:    cfg.AccrualAddr,
		log:            baseLogger,
	}
}

func (c *Controller) Run(ctx context.Context, requestInterval int) {
	ch := make(chan data.Order, 10)

	c.log.Info("[accrual:Controller:Run] start interaction with loyalty system")

	go c.requestCalculationsResult(ctx, ch, time.Duration(requestInterval))
	go c.updateOrders(ctx, ch)
}

func (c *Controller) requestCalculationsResult(ctx context.Context, chOut chan<- data.Order, requestInterval time.Duration) {
	ticker := time.NewTicker(requestInterval * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.log.Info("[accrual:Controller:requestCalculationsResult] requests task is stopping")
			close(chOut)
			return
		case <-ticker.C:
			for _, status := range []string{"PROCESSING", "NEW"} {
				orders, err := c.ordersStorage.GetOrders(ctx, map[string]interface{}{"status_id": data.GetOrderStatusID(status)})
				if err != nil {
					c.log.Info("[accrual:Controller:requestCalculationsResult] failed to get orders with PROCESSING status: %v", err)
				} else {
					for i := 0; i < len(orders); i++ {
						respStatus, pause, err := c.client.RequestCalculationResult(ctx, c.accrualAddr, &orders[i])
						if err != nil {
							if errors.Is(err, context.Canceled) {
								c.log.Info("[accrual:Controller:requestCalculationsResult] requests task is stopping: %v", err)
								close(chOut)
								return
							}

							c.log.Info("[accrual:Controller:requestCalculationsResult] failed ('%d') to get calculation from loyalty system for order '%v': %v", respStatus, orders[i], err)
							continue
						}

						if respStatus == http.StatusTooManyRequests && pause != 0 {
							c.pauseRequest(ctx, pause)
							i--
						} else {
							chOut <- orders[i]
						}
					}
				}
			}
		}
	}
}

func (c *Controller) pauseRequest(ctx context.Context, interval client.RetryInterval) {
	c.log.Info("[accrual:Controller:pauseRequest] start request pause '%d' duration", interval)
	timer := time.NewTimer(time.Duration(interval) * time.Second)

	for {
		select {
		case <-ctx.Done():
			c.log.Info("[accrual:Controller:pauseRequest] pause has been stopped by context")
			return
		case <-timer.C:
			c.log.Info("[accrual:Controller:pauseRequest] pause has been finished")
			return
		}
	}
}

func (c *Controller) updateOrders(ctx context.Context, chIn <-chan data.Order) {
	for {
		select {
		case <-ctx.Done():
			c.log.Info("[accrual:Controller:updateOrders] update orders task is stopping")
			return
		case order, ok := <-chIn:
			if !ok {
				c.log.Info("[accrual:Controller:updateOrders] stop action. channel was closed.")
				return
			}

			orderStatusID := data.GetOrderStatusID(order.Status)
			if orderStatusID > data.StatusProcessing {
				if err := c.ordersStorage.UpdateOrder(ctx, &order); err != nil {
					c.log.Info("[accrual:Controller:updateOrders] error occurred during order '%v' update in db: %v", order, err)
				}

				if orderStatusID == data.StatusProcessed {
					if err := c.bonusesStorage.AddBonuses(ctx, order.UserID, order.Accrual); err != nil {
						c.log.Info("[accrual:Controller:updateOrders] error occurred during add bonuses for order '%v' in db: %v", order, err)
					}
				}
			}
		}
	}
}
