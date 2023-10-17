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

// Insert performs direct query request to database to add new bonuses record.
func Insert(ctx context.Context, tx *sql.Tx, userID int64, count float32, log logger.BaseLogger) (int64, error) {
	errMsg := fmt.Sprintf("insert bonuses '%f' for userID '%d' in '%s'", count, userID, GetTableFullName(BonusesTable)) + ": %w"

	stmt, err := createInsertStmt(ctx, tx)
	if err != nil {
		return -1, fmt.Errorf(errMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, log)

	var bonusID int64
	query := func(context context.Context) error {
		_ = stmt.QueryRowContext(
			context,
			userID,
			count,
		).Scan(&bonusID)

		return nil
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, log, []int{1, 1, 3}, db.DatabaseErrorsToRetry, query)
	if err != nil {
		return -1, fmt.Errorf(errMsg, err)
	}

	return bonusID, nil
}

// createUpdateBonusesStmt generates statement for insert query.
func createInsertStmt(ctx context.Context, tx *sql.Tx) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	psqlInsert, _, err := psql.Insert(GetTableFullName(BonusesTable)).
		Columns(ColumnsInBonusesTable...).
		Values(make([]interface{}, len(ColumnsInBonusesTable))...).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql insert statement for '%s': %w", GetTableFullName(BonusesTable), err)
	}
	return tx.PrepareContext(ctx, psqlInsert)
}
