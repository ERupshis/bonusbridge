package managers

import (
	"context"
	"fmt"

	"github.com/erupshis/bonusbridge/internal/auth/users/data"
	"github.com/erupshis/bonusbridge/internal/db"
	"github.com/erupshis/bonusbridge/internal/db/queries/users"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
)

// manager storageManager implementation for PostgreSQL. Consist of database and QueriesHandler.
// Request to database are synchronized by sync.RWMutex. All requests are done on united transaction. Multi insert/update/delete is not supported at the moment.
type manager struct {
	*db.Conn

	log logger.BaseLogger
}

// Create creates manager implementation. Supports migrations and check connection to database.
func Create(dbConn *db.Conn, log logger.BaseLogger) BaseUsersManager {
	return &manager{
		Conn: dbConn,
		log:  log,
	}
}

func (p *manager) AddUser(ctx context.Context, user *data.User) (int64, error) {
	p.log.Info("[users:manager:AddUser] start transaction with user data '%v'", *user)
	errMsg := "add user in db: %w"
	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return -1, fmt.Errorf(errMsg, err)
	}

	err = users.Insert(ctx, tx, user, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return -1, fmt.Errorf(errMsg, err)
	}

	err = tx.Commit()
	if err != nil {
		return -1, fmt.Errorf(errMsg, err)
	}

	p.log.Info("[users:manager:AddUser] transaction successful")
	return p.GetUserID(ctx, user.Login)
}

func (p *manager) GetUser(ctx context.Context, login string) (*data.User, error) {
	user, err := p.getUser(ctx, map[string]interface{}{"login": login})
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	if user == nil {
		return nil, nil
	}

	return user, nil
}

func (p *manager) GetUserID(ctx context.Context, login string) (int64, error) {
	user, err := p.getUser(ctx, map[string]interface{}{"login": login})
	if err != nil {
		return -1, fmt.Errorf("get user ID: %w", err)
	}

	if user == nil {
		return -1, nil
	}

	return user.ID, nil
}

func (p *manager) GetUserRole(ctx context.Context, userID int64) (int, error) {
	user, err := p.getUser(ctx, map[string]interface{}{"id": userID})
	if err != nil {
		return -1, fmt.Errorf("get user role: %w", err)
	}

	if user == nil {
		return -1, nil
	}

	return user.Role, nil
}

func (p *manager) getUser(ctx context.Context, filters map[string]interface{}) (*data.User, error) {
	p.log.Info("[users:manager:getUser] start transaction with filters '%v'", filters)
	errMsg := "get user: %w"
	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	usersSelected, err := users.Select(ctx, tx, filters, p.log)
	if err != nil {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return nil, fmt.Errorf(errMsg, err)
	}

	if len(usersSelected) > 1 {
		helpers.ExecuteWithLogError(tx.Rollback, p.log)
		return nil, data.ErrUserNotFound
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	p.log.Info("[users:manager:getUser] transaction successful")

	if len(usersSelected) == 0 {
		return nil, nil
	}
	return &usersSelected[0], nil
}

//TODO: hash password in db.
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
