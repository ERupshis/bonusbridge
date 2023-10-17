// Package postgresql postgresql handling PostgreSQL database.
package postgresql

import (
	"context"
	"fmt"
	"time"

	bonusesQueries "github.com/erupshis/bonusbridge/internal/bonuses/storage/managers/postgresql/queries/bonuses"
	"github.com/erupshis/bonusbridge/internal/db"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
	"github.com/erupshis/bonusbridge/internal/orders/storage/managers"
	ordersQueries "github.com/erupshis/bonusbridge/internal/orders/storage/managers/postgresql/queries"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type manager struct {
	*db.Conn

	log logger.BaseLogger
}

// Create creates manager implementation. Supports migrations and check connection to database.
func Create(dbConn *db.Conn, log logger.BaseLogger) managers.BaseOrdersManager {
	return &manager{
		Conn: dbConn,
		log:  log,
	}
}

func (p *manager) AddOrder(ctx context.Context, number string, userID int64) (int64, error) {
	p.log.Info("[orders:manager:AddOrder] start transaction for order '%s', userID '%d'", number, userID)
	errMsg := "add order in db: %w"

	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return -1, fmt.Errorf(errMsg, err)
	}

	orders, err := ordersQueries.Select(ctx, tx, map[string]interface{}{"number": number}, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return -1, fmt.Errorf(errMsg, err)
	}

	if len(orders) != 0 {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		if orders[0].UserID == userID {
			return int64(orders[0].ID), fmt.Errorf("add order in storage: %w", data.ErrOrderWasAddedBefore)
		} else {
			return int64(orders[0].ID), fmt.Errorf("add order in storage: %w", data.ErrOrderWasAddedByAnotherUser)
		}
	}

	bonusID, err := bonusesQueries.Insert(ctx, tx, userID, 0, p.log)
	if err != nil {
		return -1, fmt.Errorf(errMsg, err)
	}

	newOrder := &data.Order{
		Number:     number,
		UserID:     userID,
		Status:     "NEW",
		BonusID:    bonusID,
		UploadedAt: time.Now(),
	}

	id, err := ordersQueries.Insert(ctx, tx, newOrder, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return -1, fmt.Errorf(errMsg, err)
	}

	err = tx.Commit()
	if err != nil {
		return -1, fmt.Errorf(errMsg, err)
	}

	p.log.Info("[orders:manager:AddOrder] transaction successful")
	return id, nil
}

func (p *manager) UpdateOrder(ctx context.Context, order *data.Order) error {
	p.log.Info("[orders:manager:UpdateOrder] start transaction for order data '%v'", *order)
	errMsg := "update order in db: %w"
	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	ordersValuesToUpdate := map[string]interface{}{
		"status_id": data.GetOrderStatusID(order.Status),
	}
	if err = ordersQueries.UpdateByID(ctx, tx, int64(order.ID), ordersValuesToUpdate, p.log); err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return fmt.Errorf(errMsg, err)
	}

	bonusesValuesToUpdate := map[string]interface{}{
		"count": order.Accrual,
	}
	if err = bonusesQueries.UpdateByID(ctx, tx, order.BonusID, bonusesValuesToUpdate, p.log); err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return fmt.Errorf(errMsg, err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	p.log.Info("[orders:manager:UpdateOrder] transaction successful")
	return nil
}

func (p *manager) GetOrders(ctx context.Context, filters map[string]interface{}) ([]data.Order, error) {
	p.log.Info("[orders:manager:GetOrders] start transaction with filters '%v'", filters)
	errMsg := "select orders in db: %w"
	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	orders, err := ordersQueries.Select(ctx, tx, filters, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return nil, fmt.Errorf(errMsg, err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	p.log.Info("[orders:manager:GetOrders] transaction successful")
	return orders, nil
}
