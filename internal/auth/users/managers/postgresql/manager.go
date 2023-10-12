package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/erupshis/bonusbridge/internal/auth/users/managers"
	"github.com/erupshis/bonusbridge/internal/config"
	"github.com/erupshis/bonusbridge/internal/logger"
	dbData "github.com/erupshis/bonusbridge/internal/orders/storage/managers/postgresql/data"
	"github.com/erupshis/bonusbridge/internal/retryer"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

// postgresDB storageManager implementation for PostgreSQL. Consist of database and QueriesHandler.
// Request to database are synchronized by sync.RWMutex. All requests are done on united transaction. Multi insert/update/delete is not supported at the moment.
type postgresDB struct {
	database *sql.DB

	log logger.BaseLogger
	mu  sync.RWMutex
}

// CreateUsersPostgreDB creates manager implementation. Supports migrations and check connection to database.
func CreateUsersPostgreDB(ctx context.Context, cfg config.Config, log logger.BaseLogger) (managers.BaseUsersManager, error) {
	log.Info("[CreateUsersPostgreDB] open database with settings: '%s'", cfg.DatabaseDSN)
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

	log.Info("[CreateUsersPostgreDB] successful")
	return manager, nil
}

// CheckConnection checks connection to database.
func (p *postgresDB) CheckConnection(ctx context.Context) (bool, error) {
	exec := func(context context.Context) (int64, []byte, error) {
		return 0, []byte{}, p.database.PingContext(context)
	}
	_, _, err := retryer.RetryCallWithTimeout(ctx, p.log, nil, dbData.DatabaseErrorsToRetry, exec)
	if err != nil {
		return false, fmt.Errorf("check connection: %w", err)
	}
	return true, nil
}

// Close closes database.
func (p *postgresDB) Close() error {
	return p.database.Close()
}

func (p *postgresDB) AddUser(login string, password string) (int64, error) {
	return -1, nil
}

func (p *postgresDB) GetUserID(login string) (int64, error) {
	return -1, nil
}

func (p *postgresDB) GetUserRole(userID int64) (int, error) {
	return -1, nil
}

func (p *postgresDB) ValidateUser(login string, password string) (bool, error) {
	return false, nil
}

//password := "user_password" // Replace with the actual password provided by the user
//
//// Hash and salt the password
//hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
//if err != nil {
//log.Fatal(err)
//}
//
//// Store 'hashedPassword' in the database for the user
//
//// User login: Verify password
//providedPassword := "user_password" // Replace with the password provided during login
//
//// Verify the provided password with the stored hashed password
//err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(providedPassword))
//if err == nil {
//fmt.Println("Password is correct!")
//} else if err == bcrypt.ErrMismatchedHashAndPassword {
//fmt.Println("Password is incorrect.")
//} else {
//log.Fatal(err)
//}
