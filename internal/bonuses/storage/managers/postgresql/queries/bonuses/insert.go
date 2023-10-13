package bonuses

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage/managers/postgresql/data"
	"github.com/erupshis/bonusbridge/internal/dberrors"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/retryer"
)

// Insert performs direct query request to database to add new bonuses record.
func Insert(ctx context.Context, tx *sql.Tx, userID int64, count float32, log logger.BaseLogger) error {
	errMsg := fmt.Sprintf("insert bonuses '%f' for userID '%d' in '%s'", count, userID, data.GetTableFullName(data.BonusesTable)) + ": %w"

	stmt, err := createInsertBonusesStmt(ctx, tx)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, log)

	query := func(context context.Context) error {
		_, err := stmt.ExecContext(
			context,
			userID,
			count,
			0,
		)

		return err
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, log, []int{1, 1, 3}, dberrors.DatabaseErrorsToRetry, query)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	return nil
}

// createUpdateBonusesStmt generates statement for insert query.
func createInsertBonusesStmt(ctx context.Context, tx *sql.Tx) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	psqlInsert, _, err := psql.Insert(data.GetTableFullName(data.BonusesTable)).
		Columns(data.ColumnsInBonusesTable...).
		Values(make([]interface{}, len(data.ColumnsInBonusesTable))...).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql insert statement for '%s': %w", data.GetTableFullName(data.BonusesTable), err)
	}
	return tx.PrepareContext(ctx, psqlInsert)
}
