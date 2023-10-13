package queries

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/erupshis/bonusbridge/internal/auth/users/data"
	dbUsersData "github.com/erupshis/bonusbridge/internal/auth/users/managers/postgresql/data"
	"github.com/erupshis/bonusbridge/internal/dberrors"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/retryer"
)

// SelectUsers performs direct query request to database to select users satisfying filters.
func SelectUsers(ctx context.Context, tx *sql.Tx, filters map[string]interface{}, log logger.BaseLogger) ([]data.User, error) {
	errorMsg := fmt.Sprintf("select orders with filter '%v' in '%s'", filters, dbUsersData.UsersTable) + ": %w"

	stmt, err := createSelectUsersStmt(ctx, tx, filters)
	if err != nil {
		return nil, fmt.Errorf(errorMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, log)

	var valuesToUpdate []interface{}
	for _, val := range filters {
		valuesToUpdate = append(valuesToUpdate, val)
	}

	var rows *sql.Rows
	query := func(context context.Context) error {
		rows, err = stmt.QueryContext(
			context,
			valuesToUpdate...,
		)

		if err == nil {
			if rows.Err() != nil {
				return fmt.Errorf(errorMsg, rows.Err())
			}
		}

		return err
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, log, []int{1, 1, 3}, dberrors.DatabaseErrorsToRetry, query)
	if err != nil {
		return nil, fmt.Errorf(errorMsg, err)
	}

	defer helpers.ExecuteWithLogError(rows.Close, log)
	var res []data.User
	for rows.Next() {
		user := data.User{}
		err := rows.Scan(
			&user.ID,
			&user.Login,
			&user.Password,
			&user.Role,
		)
		if err != nil {
			return nil, fmt.Errorf("parse db result: %w", err)
		}

		res = append(res, user)
	}

	return res, nil
}

// createSelectUsersStmt generates statement for select query.
func createSelectUsersStmt(ctx context.Context, tx *sql.Tx, filters map[string]interface{}) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.Select(
		"id",
		"login",
		"password",
		"role_id",
	).
		From(dbUsersData.GetTableFullName(dbUsersData.UsersTable))
	if len(filters) != 0 {
		for key := range filters {
			builder = builder.Where(sq.Eq{key: "?"})
		}
	}
	psqlSelect, _, err := builder.ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql select statement for '%s': %w", dbUsersData.GetTableFullName(dbUsersData.UsersTable), err)
	}
	return tx.PrepareContext(ctx, psqlSelect)
}
