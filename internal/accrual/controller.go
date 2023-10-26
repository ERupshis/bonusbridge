package accrual

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/erupshis/bonusbridge/internal/accrual/client"
	"github.com/erupshis/bonusbridge/internal/accrual/workerspool"
	bonusesStorage "github.com/erupshis/bonusbridge/internal/bonuses/storage"
	"github.com/erupshis/bonusbridge/internal/config"
	"github.com/erupshis/bonusbridge/internal/db/queries"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
	ordersStorage "github.com/erupshis/bonusbridge/internal/orders/storage"
)

type Controller struct {
	ordersStorage  ordersStorage.BaseOrdersStorage
	bonusesStorage bonusesStorage.BaseBonusesStorage

	client client.BaseClient

	workersPool *workerspool.Pool

	accrualAddr string

	log logger.BaseLogger
}

func CreateController(ordersStorage ordersStorage.BaseOrdersStorage,
	bonusesStorage bonusesStorage.BaseBonusesStorage,
	client client.BaseClient,
	workersPool *workerspool.Pool,
	cfg config.Config,
	baseLogger logger.BaseLogger) Controller {
	return Controller{
		ordersStorage:  ordersStorage,
		bonusesStorage: bonusesStorage,
		client:         client,
		workersPool:    workersPool,
		accrualAddr:    cfg.AccrualAddr,
		log:            baseLogger,
	}
}

func (c *Controller) Run(ctx context.Context, requestInterval int) {
	c.log.Info("[accrual:Controller:Run] start interaction with loyalty system, requests interval '%d' seconds", requestInterval)

	go c.requestCalculationsResult(ctx, time.Duration(requestInterval))
	go c.updateOrders(ctx)
}

func (c *Controller) requestCalculationsResult(ctx context.Context, requestInterval time.Duration) {
	ticker := time.NewTicker(requestInterval * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.log.Info("[accrual:Controller:requestCalculationsResult] requests task is stopping by context")
			return
		case <-ticker.C:
			c.processOrders(ctx, c.workersPool)
		}
	}
}

func (c *Controller) processOrders(ctx context.Context, workersPool *workerspool.Pool) {
	orders, err := c.ordersStorage.GetOrders(ctx, map[string]interface{}{queries.Custom: fmt.Sprintf("orders.status_id <= %d", data.StatusInvalid)})
	if err != nil {
		c.log.Info("[accrual:Controller:requestCalculationsResult] failed to get orders with PROCESSING status: %v", err)
		return
	}

	for i := 0; i < len(orders); i++ {
		c.addJobForWorkers(ctx, workersPool, orders[i])
	}
}

func (c *Controller) addJobForWorkers(ctx context.Context, workersPool *workerspool.Pool, order data.Order) {
	workersPool.AddJob(func() (*data.Order, error) {
		respStatus, pause, err := c.client.RequestCalculationResult(ctx, c.accrualAddr, &order)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				c.log.Info("[accrual:Controller:requestCalculationsResult] requests task was interrupted: %v", err)
				return nil, nil
			}

			c.log.Info("[accrual:Controller:requestCalculationsResult] failed ('%d') to get calculation from loyalty system for order '%v': %v", respStatus, order, err)
			return nil, fmt.Errorf("request to accrual system: %w", err)
		}

		if needPauseRequests(respStatus, pause) {
			c.pauseRequest(ctx, pause)
			return nil, fmt.Errorf("request skipped, accrual was overload: %w", err)
		} else if respStatus == http.StatusOK {
			return &order, nil
		}

		return nil, fmt.Errorf("request to accrual finished with status '%d' and error: %v", respStatus, err)
	})
}

func needPauseRequests(respStatus client.ResponseStatus, pause client.RetryInterval) bool {
	return respStatus == http.StatusTooManyRequests && pause != 0
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

func (c *Controller) updateOrders(ctx context.Context) {
	chIn := c.workersPool.GetResultChan()

	for {
		select {
		case <-ctx.Done():
			c.log.Info("[accrual:Controller:updateOrders] update orders task is stopping by context")
			return
		case order, ok := <-chIn:
			if !ok {
				c.log.Info("[accrual:Controller:updateOrders] stop action. channel was closed.")
				return
			}

			orderStatusID := data.GetOrderStatusID(order.Status)
			if orderStatusID > data.StatusProcessing {
				if err := c.ordersStorage.UpdateOrder(ctx, order); err != nil {
					c.log.Info("[accrual:Controller:updateOrders] error occurred during order '%v' update in db: %v", order, err)
				}
			}
		}
	}
}
