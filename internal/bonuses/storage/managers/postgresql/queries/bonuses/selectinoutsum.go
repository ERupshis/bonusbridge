package bonuses

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	dbBonusesData "github.com/erupshis/bonusbridge/internal/bonuses/storage/managers/postgresql/data"
	"github.com/erupshis/bonusbridge/internal/db"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/retryer"
)

func SelectInOutSumByUserID(ctx context.Context, tx *sql.Tx, in bool, userID int64, log logger.BaseLogger) (float32, error) {
	errMsg := fmt.Sprintf("select bonuses balance for userID '%d' in '%s'", userID, dbBonusesData.GetTableFullName(dbBonusesData.BonusesTable)) + ": %w"

	var stmt *sql.Stmt
	var err error
	if in {
		stmt, err = createSelectInSumByUserIDStmt(ctx, tx)
	} else {
		stmt, err = createSelectOutSumByUserIDStmt(ctx, tx)
	}

	if err != nil {
		return -1.0, fmt.Errorf(errMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, log)

	var rows *sql.Rows
	query := func(context context.Context) error {
		rows, err = stmt.QueryContext(
			context,
			userID,
			0,
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
		return -1.0, fmt.Errorf(errMsg, err)
	}

	defer helpers.ExecuteWithLogError(rows.Close, log)
	var res float32
	for rows.Next() {
		err = rows.Scan(
			&res,
		)
		if err != nil {
			return -1.0, fmt.Errorf("parse db result: %w", err)
		}
	}

	return res, nil
}

func createSelectInSumByUserIDStmt(ctx context.Context, tx *sql.Tx) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	psqlSelect, _, err := psql.Select("SUM(count)").
		From(dbBonusesData.GetTableFullName(dbBonusesData.BonusesTable)).
		Where(sq.Eq{"user_id": 0}).
		Where(sq.GtOrEq{"count": 0}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql select statement for '%s': %w", dbBonusesData.GetTableFullName(dbBonusesData.BonusesTable), err)
	}
	return tx.PrepareContext(ctx, psqlSelect)
}

func createSelectOutSumByUserIDStmt(ctx context.Context, tx *sql.Tx) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	psqlSelect, _, err := psql.Select("SUM(count)").
		From(dbBonusesData.GetTableFullName(dbBonusesData.BonusesTable)).
		Where(sq.Eq{"user_id": 0}).
		Where(sq.LtOrEq{"count": 0}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql select statement for '%s': %w", dbBonusesData.GetTableFullName(dbBonusesData.BonusesTable), err)
	}
	return tx.PrepareContext(ctx, psqlSelect)
}
