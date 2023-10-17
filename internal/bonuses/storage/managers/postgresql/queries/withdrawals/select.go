package withdrawals

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

// Select performs direct query request to database to select withdrawals satisfying filters.
func Select(ctx context.Context, tx *sql.Tx, filters map[string]interface{}, log logger.BaseLogger) ([]data.Withdrawal, error) {
	errMsg := fmt.Sprintf("select withdrawals with filter '%v' in '%s'",
		filters,
		dbBonusesData.GetTableFullName(dbBonusesData.WithdrawalsTable),
	) + ": %w"

	stmt, err := createSelectWithdrawalsStmt(ctx, tx, filters)
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
	var res []data.Withdrawal
	for rows.Next() {
		withdrawal := data.Withdrawal{}
		err := rows.Scan(
			&withdrawal.ID,
			&withdrawal.UserID,
			&withdrawal.Order,
			&withdrawal.Sum,
			&withdrawal.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("parse db result: %w", err)
		}

		res = append(res, withdrawal)
	}

	return res, nil
}

// createSelectBonusesStmt generates statement for select query.
func createSelectWithdrawalsStmt(ctx context.Context, tx *sql.Tx, filters map[string]interface{}) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	bonusesJoin := fmt.Sprintf("LEFT JOIN %s ON %[1]s.id = %s.bonus_id",
		dbBonusesData.GetTableFullName(dbBonusesData.BonusesTable),
		dbBonusesData.GetTableFullName(dbBonusesData.WithdrawalsTable),
	)

	builder := psql.Select(
		dbBonusesData.GetTableFullName(dbBonusesData.WithdrawalsTable)+".id",
		dbBonusesData.GetTableFullName(dbBonusesData.WithdrawalsTable)+".user_id",
		dbBonusesData.GetTableFullName(dbBonusesData.WithdrawalsTable)+".order_num",
		fmt.Sprintf("ABS(%s) AS sum", dbBonusesData.GetTableFullName(dbBonusesData.BonusesTable)+".count"),
		dbBonusesData.GetTableFullName(dbBonusesData.WithdrawalsTable)+".processed_at",
	).
		From(dbBonusesData.GetTableFullName(dbBonusesData.WithdrawalsTable)).
		JoinClause(bonusesJoin)

	for key := range filters {
		switch key {
		case "user_id":
			key = dbBonusesData.GetTableFullName(dbBonusesData.WithdrawalsTable) + ".user_id"
		}
		builder = builder.Where(sq.Eq{key: "?"})
	}
	psqlSelect, _, err := builder.ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql select statement for '%s': %w", dbBonusesData.GetTableFullName(dbBonusesData.WithdrawalsTable), err)
	}
	return tx.PrepareContext(ctx, psqlSelect)
}
