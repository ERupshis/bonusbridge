package bonuses

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/erupshis/bonusbridge/internal/db"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/retryer"
)

const (
	SumTotal = iota
	SumIn
	SumOut
)

func SelectSumByUserID(ctx context.Context, tx *sql.Tx, filter int, userID int64, log logger.BaseLogger) (float32, error) {
	errMsg := fmt.Sprintf("select bonuses balance for userID '%d' in '%s'", userID, BonusesTable) + ": %w"

	stmt, err := createSelectSumByUserIDStmt(ctx, tx, filter)
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
	var res sql.NullFloat64
	for rows.Next() {
		err = rows.Scan(
			&res,
		)
		if err != nil {
			return -1.0, fmt.Errorf("parse db result: %w", err)
		}
	}

	return float32(res.Float64), nil
}

func createSelectSumByUserIDStmt(ctx context.Context, tx *sql.Tx, filter int) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.Select("SUM(count)").
		From(BonusesTable).
		Where(sq.Eq{"user_id": 0})

	switch filter {
	case SumIn:
		builder = builder.Where(sq.GtOrEq{"count": 0})
	case SumOut:
		builder = builder.Where(sq.LtOrEq{"count": 0})
	default:
		builder = builder.Where(sq.GtOrEq{"id": 0})
	}

	psqlSelect, _, err := builder.ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql select statement for '%s': %w", BonusesTable, err)
	}
	return tx.PrepareContext(ctx, psqlSelect)
}
