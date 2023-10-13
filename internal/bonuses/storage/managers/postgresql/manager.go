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
	"github.com/erupshis/bonusbridge/internal/config"
	"github.com/erupshis/bonusbridge/internal/dberrors"
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
	mu       sync.RWMutex
	database *sql.DB

	log logger.BaseLogger
}

// CreateBonusesPostgreDB creates manager implementation. Supports migrations and check connection to database.
func CreateBonusesPostgreDB(ctx context.Context, cfg config.Config, log logger.BaseLogger) (managers.BaseBonusesManager, error) {
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
	return nil
}

func (p *postgresDB) GetBonuses(ctx context.Context, userID int64) (float32, error) {
	return -1, nil
}

func (p *postgresDB) WithdrawBonuses(ctx context.Context, userID int64, count float32) (bool, error) {
	return false, nil
}

func (p *postgresDB) GetWithdrawnBonuses(ctx context.Context, userID int64) (float32, error) {
	return -1, nil
}

func (p *postgresDB) GetWithdrawals(ctx context.Context, userID int64) ([]data.Withdrawal, error) {
	return nil, nil
}
