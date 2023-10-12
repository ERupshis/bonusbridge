package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/erupshis/bonusbridge/internal/auth/users/data"
	"github.com/erupshis/bonusbridge/internal/auth/users/managers"
	"github.com/erupshis/bonusbridge/internal/auth/users/managers/postgresql/queries"
	"github.com/erupshis/bonusbridge/internal/config"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	dbData "github.com/erupshis/bonusbridge/internal/orders/storage/managers/postgresql/data"
	"github.com/erupshis/bonusbridge/internal/retryer"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
)

// postgresDB storageManager implementation for PostgreSQL. Consist of database and QueriesHandler.
// Request to database are synchronized by sync.RWMutex. All requests are done on united transaction. Multi insert/update/delete is not supported at the moment.
type postgresDB struct {
	mu       sync.RWMutex
	database *sql.DB

	log logger.BaseLogger
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

func (p *postgresDB) AddUser(ctx context.Context, user *data.User) (int64, error) {
	p.mu.Lock()

	p.log.Info("[users:postgresDB:AddUser] start transaction")
	errMsg := "add user in db: %w"
	tx, err := p.database.BeginTx(ctx, nil)
	if err != nil {
		return -1, fmt.Errorf(errMsg, err)
	}

	err = queries.InsertUser(ctx, tx, user, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return -1, fmt.Errorf(errMsg, err)
	}

	err = tx.Commit()
	if err != nil {
		return -1, fmt.Errorf(errMsg, err)
	}

	p.log.Info("[users:postgresDB:AddUser] transaction successful")
	p.mu.Unlock()
	return p.GetUserID(ctx, user.Login)
}

func (p *postgresDB) GetUser(ctx context.Context, login string) (*data.User, error) {
	user, err := p.getUser(ctx, map[string]interface{}{"login": login})
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	if user == nil {
		return nil, nil
	}

	return user, nil
}

func (p *postgresDB) GetUserID(ctx context.Context, login string) (int64, error) {
	user, err := p.getUser(ctx, map[string]interface{}{"login": login})
	if err != nil {
		return -1, fmt.Errorf("get user ID: %w", err)
	}

	if user == nil {
		return -1, nil
	}

	return user.ID, nil
}

func (p *postgresDB) GetUserRole(ctx context.Context, userID int64) (int, error) {
	user, err := p.getUser(ctx, map[string]interface{}{"id": userID})
	if err != nil {
		return -1, fmt.Errorf("get user role: %w", err)
	}

	if user == nil {
		return -1, nil
	}

	return user.Role, nil
}

//func (p *postgresDB) ValidateUser(ctx context.Context, login string, password string) (bool, error) {
//	user, err := p.getUser(ctx, map[string]interface{}{"login": login})
//	if err != nil {
//		return false, fmt.Errorf("validate user: %w", err)
//	}
//
//	if user == nil {
//		return false, fmt.Errorf("validate user: user not found")
//	}
//
//	return password == user.Password, nil
//}

func (p *postgresDB) getUser(ctx context.Context, filters map[string]interface{}) (*data.User, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	p.log.Info("[users:postgresDB:getUser] start transaction")
	errMsg := "get user: %w"
	tx, err := p.database.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	users, err := queries.SelectUsers(ctx, tx, filters, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return nil, fmt.Errorf(errMsg, err)
	}

	if len(users) > 1 {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return nil, fmt.Errorf("user is not found in db or few users has the same login")
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	p.log.Info("[users:postgresDB:getUser] transaction successful")

	if len(users) == 0 {
		return nil, nil
	}
	return &users[0], nil
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
