// Package postgresql postgresql handling PostgreSQL database.
package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/erupshis/bonusbridge/internal/config"
	"github.com/erupshis/bonusbridge/internal/dberrors"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
	"github.com/erupshis/bonusbridge/internal/orders/storage/managers"
	"github.com/erupshis/bonusbridge/internal/orders/storage/managers/postgresql/queries"
	"github.com/erupshis/bonusbridge/internal/retryer"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// postgresDB storageManager implementation for PostgreSQL. Consist of database and QueriesHandler.
// Request to database are synchronized by sync.RWMutex. All requests are done on united transaction. Multi insert/update/delete is not supported at the moment.
type postgresDB struct {
	database *sql.DB

	log logger.BaseLogger
}

// Create creates manager implementation. Supports migrations and check connection to database.
func Create(ctx context.Context, cfg config.Config, log logger.BaseLogger) (managers.BaseOrdersManager, error) {
	log.Info("[orders:postgresDB:Create] open database with settings: '%s'", cfg.DatabaseDSN)
	createDatabaseError := "create db: %w"
	database, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf(createDatabaseError, err)
	}

	driver, err := postgres.WithInstance(database, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf(createDatabaseError, err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://db/migrations/", "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf(createDatabaseError, err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf(createDatabaseError, err)
	}

	manager := &postgresDB{
		database: database,
		log:      log,
	}

	if _, err = manager.CheckConnection(ctx); err != nil {
		return nil, fmt.Errorf(createDatabaseError, err)
	}

	log.Info("[orders:postgresDB:Create] successful")
	return manager, nil
}

// CheckConnection checks connection to database.
func (p *postgresDB) CheckConnection(ctx context.Context) (bool, error) {
	exec := func(context context.Context) (int64, []byte, error) {
		return 0, []byte{}, p.database.PingContext(context)
	}
	_, _, err := retryer.RetryCallWithTimeout(ctx, p.log, nil, dberrors.DatabaseErrorsToRetry, exec)
	if err != nil {
		return false, fmt.Errorf("check connection: %w", err)
	}
	return true, nil
}

// Close closes database.
func (p *postgresDB) Close() error {
	return p.database.Close()
}

func (p *postgresDB) AddOrder(ctx context.Context, number string, userID int64) (int64, error) {
	newOrder := &data.Order{
		Number:     number,
		UserID:     userID,
		Status:     "NEW",
		Accrual:    0.0,
		UploadedAt: time.Now(),
	}

	p.log.Info("[orders:postgresDB:AddOrder] start transaction for order '%s', userID '%d'", number, userID)
	errMsg := "add order in db: %w"
	tx, err := p.database.BeginTx(ctx, nil)
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

	p.log.Info("[orders:postgresDB:AddOrder] transaction successful")
	return id, nil
}

func (p *postgresDB) UpdateOrder(ctx context.Context, order *data.Order) error {
	p.log.Info("[orders:postgresDB:UpdateOrder] start transaction for order data '%v'", *order)
	errMsg := "update order in db: %w"
	tx, err := p.database.BeginTx(ctx, nil)
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

	p.log.Info("[orders:postgresDB:UpdateOrder] transaction successful")
	return nil
}

func (p *postgresDB) GetOrder(ctx context.Context, number string) (*data.Order, error) {
	p.log.Info("[orders:postgresDB:GetOrder] start transaction for order number '%s'", number)
	errMsg := "select order in db: %w"
	tx, err := p.database.BeginTx(ctx, nil)
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

	p.log.Info("[orders:postgresDB:GetOrder] transaction successful")
	if len(orders) == 0 {
		return nil, nil
	} else if len(orders) > 1 {
		return nil, fmt.Errorf(errMsg, fmt.Errorf("more than one order in db with number '%s'", number))
	}

	return &orders[0], nil
}
func (p *postgresDB) GetOrders(ctx context.Context, filters map[string]interface{}) ([]data.Order, error) {
	p.log.Info("[orders:postgresDB:GetOrders] start transaction with filters '%v'", filters)
	errMsg := "select orders in db: %w"
	tx, err := p.database.BeginTx(ctx, nil)
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

	p.log.Info("[orders:postgresDB:GetOrders] transaction successful")
	return orders, nil
}
