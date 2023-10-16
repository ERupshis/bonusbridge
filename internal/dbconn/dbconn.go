package dbconn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/erupshis/bonusbridge/internal/config"
	"github.com/erupshis/bonusbridge/internal/dberrors"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/retryer"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const migrationsFolder = "file://db/migrations/"

// DBConn storageManager implementation for PostgreSQL. Consist of database.
type DBConn struct {
	*sql.DB
	log logger.BaseLogger
}

// Create creates manager implementation. Supports migrations and check connection to database.
func Create(ctx context.Context, cfg config.Config, log logger.BaseLogger) (*DBConn, error) {
	log.Info("[dbconn:Create] open database with settings: '%s'", cfg.DatabaseDSN)
	createDatabaseError := "create db: %w"
	database, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf(createDatabaseError, err)
	}

	driver, err := postgres.WithInstance(database, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf(createDatabaseError, err)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsFolder, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf(createDatabaseError, err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf(createDatabaseError, err)
	}

	manager := &DBConn{
		DB:  database,
		log: log,
	}

	if _, err = manager.CheckConnection(ctx); err != nil {
		return nil, fmt.Errorf(createDatabaseError, err)
	}

	log.Info("[dbconn:Create] successful")
	return manager, nil
}

// CheckConnection checks connection to database.
func (p *DBConn) CheckConnection(ctx context.Context) (bool, error) {
	exec := func(context context.Context) (int64, []byte, error) {
		return 0, []byte{}, p.PingContext(context)
	}
	_, _, err := retryer.RetryCallWithTimeout(ctx, p.log, nil, dberrors.DatabaseErrorsToRetry, exec)
	if err != nil {
		return false, fmt.Errorf("check connection: %w", err)
	}
	return true, nil
}

// Close closes database.
func (p *DBConn) Close() error {
	return p.DB.Close()
}
