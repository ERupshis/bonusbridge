package queries

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/erupshis/bonusbridge/internal/auth/users/data"
	dbUsersData "github.com/erupshis/bonusbridge/internal/auth/users/managers/postgresql/data"
	"github.com/erupshis/bonusbridge/internal/db"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/retryer"
)

// InsertUser performs direct query request to database to add new user.
func InsertUser(ctx context.Context, tx *sql.Tx, userData *data.User, log logger.BaseLogger) error {
	errMsg := fmt.Sprintf("insert user '%v' in '%s'", *userData, dbUsersData.GetTableFullName(dbUsersData.UsersTable)) + ": %w"

	stmt, err := createInsertUserStmt(ctx, tx)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, log)

	query := func(context context.Context) error {
		_, err := stmt.ExecContext(
			context,
			userData.Login,
			userData.Password,
			userData.Role,
		)

		return err
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, log, []int{1, 1, 3}, db.DatabaseErrorsToRetry, query)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	return nil
}

// createInsertUserStmt generates statement for insert query.
func createInsertUserStmt(ctx context.Context, tx *sql.Tx) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	psqlInsert, _, err := psql.Insert(dbUsersData.GetTableFullName(dbUsersData.UsersTable)).
		Columns(dbUsersData.ColumnsInUsersTable...).
		Values(make([]interface{}, len(dbUsersData.ColumnsInUsersTable))...).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql insert statement for '%s': %w", dbUsersData.GetTableFullName(dbUsersData.UsersTable), err)
	}
	return tx.PrepareContext(ctx, psqlInsert)
}
