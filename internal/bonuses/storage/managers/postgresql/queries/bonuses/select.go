package bonuses

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/erupshis/bonusbridge/internal/bonuses/data"
	dbBonusesData "github.com/erupshis/bonusbridge/internal/bonuses/storage/managers/postgresql/data"
	"github.com/erupshis/bonusbridge/internal/db"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/retryer"
)

// Select performs direct query request to database to select bonuses satisfying filters.
func Select(ctx context.Context, tx *sql.Tx, filters map[string]interface{}, log logger.BaseLogger) ([]data.Balance, error) {
	errMsg := fmt.Sprintf("select orders with filter '%v' in '%s'", filters, dbBonusesData.GetTableFullName(dbBonusesData.BonusesTable)) + ": %w"

	stmt, err := createSelectBonusesStmt(ctx, tx, filters)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
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
				return fmt.Errorf(errMsg, rows.Err())
			}
		}

		return err
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, log, []int{1, 1, 3}, db.DatabaseErrorsToRetry, query)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	defer helpers.ExecuteWithLogError(rows.Close, log)
	var res []data.Balance
	for rows.Next() {
		order := data.Balance{}
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Current,
			&order.Withdrawn,
		)
		if err != nil {
			return nil, fmt.Errorf("parse db result: %w", err)
		}

		res = append(res, order)
	}

	return res, nil
}

// createSelectBonusesStmt generates statement for select query.
func createSelectBonusesStmt(ctx context.Context, tx *sql.Tx, filters map[string]interface{}) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.Select(
		"id",
		"user_id",
		"balance",
		"withdrawn",
	).
		From(dbBonusesData.GetTableFullName(dbBonusesData.BonusesTable))

	for key := range filters {
		builder = builder.Where(sq.Eq{key: "?"})
	}
	psqlSelect, _, err := builder.ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql select statement for '%s': %w", dbBonusesData.GetTableFullName(dbBonusesData.BonusesTable), err)
	}
	return tx.PrepareContext(ctx, psqlSelect)
}
