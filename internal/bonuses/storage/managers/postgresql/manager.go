// Package postgresql postgresql handling PostgreSQL database.
package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/erupshis/bonusbridge/internal/bonuses/data"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage/managers"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage/managers/postgresql/queries/bonuses"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage/managers/postgresql/queries/withdrawals"
	"github.com/erupshis/bonusbridge/internal/config"
	"github.com/erupshis/bonusbridge/internal/dberrors"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/retryer"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// postgresDB storageManager implementation for PostgreSQL. Consist of database.
// Request to database are synchronized by sync.RWMutex. All requests are done on united transaction. Multi insert/update/delete is not supported at the moment.
type postgresDB struct {
	mu       *sync.RWMutex
	database *sql.DB

	log logger.BaseLogger
}

// CreateBonusesPostgreDB creates manager implementation. Supports migrations and check connection to database.
func CreateBonusesPostgreDB(ctx context.Context, cfg config.Config, mu *sync.RWMutex, log logger.BaseLogger) (managers.BaseBonusesManager, error) {
	log.Info("[CreateBonusesPostgreDB] open database with settings: '%s'", cfg.DatabaseDSN)
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
		mu:       mu,
		log:      log,
	}

	if _, err = manager.CheckConnection(ctx); err != nil {
		return nil, fmt.Errorf(createDatabaseError, err)
	}

	log.Info("[CreateBonusesPostgreDB] successful")
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

func (p *postgresDB) AddBonuses(ctx context.Context, userID int64, count float32) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	errMsg := "add bonuses in db: %w"
	p.log.Info("[bonuses:postgresDB:AddBonuses] start transaction for userID '%d', count '%d'", userID, count)
	tx, err := p.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	bonusesArr, err := bonuses.Select(ctx, tx, map[string]interface{}{"user_id": userID}, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return fmt.Errorf(errMsg, err)
	}

	if len(bonusesArr) == 0 {
		if err = bonuses.Insert(ctx, tx, userID, count, p.log); err != nil {
			helpers.ExecuteWithLogError(tx.Rollback, p.log)
			return fmt.Errorf(errMsg, err)
		}
	} else {
		if err = bonuses.UpdateByID(ctx, tx, bonusesArr[0].ID, map[string]interface{}{"balance": bonusesArr[0].Current + count}, p.log); err != nil {
			helpers.ExecuteWithLogError(tx.Rollback, p.log)
			return fmt.Errorf(errMsg, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	p.log.Info("[bonuses:postgresDB:AddBonuses] transaction successful")
	return nil
}

func (p *postgresDB) GetBalance(ctx context.Context, userID int64) (*data.Balance, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	p.log.Info("[bonuses:postgresDB:GetBalance] start transaction for userID '%d'", userID)
	errMsg := "get bonuses balance in db: %w"
	tx, err := p.database.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	bonusesArr, err := bonuses.Select(ctx, tx, map[string]interface{}{"user_id": userID}, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return nil, fmt.Errorf(errMsg, err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	p.log.Info("[bonuses:postgresDB:GetBalance] transaction successful")
	if len(bonusesArr) == 0 {
		return nil, nil
	}
	return &bonusesArr[0], nil
}

func (p *postgresDB) WithdrawBonuses(ctx context.Context, withdrawal *data.Withdrawal) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.log.Info("[bonuses:postgresDB:WithdrawBonuses] start transaction for withdrawal '%v'", *withdrawal)
	errMsg := "withdraw bonuses in db: %w"
	tx, err := p.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	bonusesArr, err := bonuses.Select(ctx, tx, map[string]interface{}{"user_id": withdrawal.UserID}, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return fmt.Errorf(errMsg, err)
	}

	if len(bonusesArr) == 0 {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return fmt.Errorf(errMsg, fmt.Errorf("bonuses record wasn't found for userID '%d'", withdrawal.UserID))
	}

	bonusesArr[0].Current -= withdrawal.Sum
	bonusesArr[0].Withdrawn += withdrawal.Sum

	valuesToUpdate := map[string]interface{}{
		"balance":   bonusesArr[0].Current,
		"withdrawn": bonusesArr[0].Withdrawn,
	}

	if err = bonuses.UpdateByID(ctx, tx, bonusesArr[0].ID, valuesToUpdate, p.log); err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return fmt.Errorf(errMsg, err)
	}

	if err = withdrawals.Insert(ctx, tx, withdrawal, p.log); err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return fmt.Errorf(errMsg, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	p.log.Info("[bonuses:postgresDB:WithdrawBonuses] transaction successful")
	return nil
}

func (p *postgresDB) GetWithdrawals(ctx context.Context, userID int64) ([]data.Withdrawal, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	p.log.Info("[bonuses:postgresDB:GetWithdrawals] start transaction for userID '%d'", userID)
	errMsg := "get withdrawals from db: %w"
	tx, err := p.database.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	withdrawalsArr, err := withdrawals.Select(ctx, tx, map[string]interface{}{"user_id": userID}, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return nil, fmt.Errorf(errMsg, err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	p.log.Info("[bonuses:postgresDB:GetWithdrawals] transaction successful")
	return withdrawalsArr, nil
}
