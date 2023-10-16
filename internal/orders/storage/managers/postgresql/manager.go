// Package postgresql postgresql handling PostgreSQL database.
package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/erupshis/bonusbridge/internal/db"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
	"github.com/erupshis/bonusbridge/internal/orders/storage/managers"
	"github.com/erupshis/bonusbridge/internal/orders/storage/managers/postgresql/queries"
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
	newOrder := &data.Order{
		Number:     number,
		UserID:     userID,
		Status:     "NEW",
		Accrual:    0.0,
		UploadedAt: time.Now(),
	}

	p.log.Info("[orders:manager:AddOrder] start transaction for order '%s', userID '%d'", number, userID)
	errMsg := "add order in db: %w"
	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return -1, fmt.Errorf(errMsg, err)
	}

	id, err := queries.InsertOrder(ctx, tx, newOrder, p.log)
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

	valuesToUpdate := map[string]interface{}{
		"status_id": data.GetOrderStatusID(order.Status),
		"accrual":   order.Accrual,
	}

	err = queries.UpdateByID(ctx, tx, int64(order.ID), valuesToUpdate, p.log)
	if err != nil {
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

func (p *manager) GetOrder(ctx context.Context, number string) (*data.Order, error) {
	p.log.Info("[orders:manager:GetOrder] start transaction for order number '%s'", number)
	errMsg := "select order in db: %w"
	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	orders, err := queries.SelectOrders(ctx, tx, map[string]interface{}{"number": number}, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return nil, fmt.Errorf(errMsg, err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	p.log.Info("[orders:manager:GetOrder] transaction successful")
	if len(orders) == 0 {
		return nil, nil
	} else if len(orders) > 1 {
		return nil, fmt.Errorf(errMsg, fmt.Errorf("more than one order in db with number '%s'", number))
	}

	return &orders[0], nil
}
func (p *manager) GetOrders(ctx context.Context, filters map[string]interface{}) ([]data.Order, error) {
	p.log.Info("[orders:manager:GetOrders] start transaction with filters '%v'", filters)
	errMsg := "select orders in db: %w"
	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	orders, err := queries.SelectOrders(ctx, tx, filters, p.log)
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
