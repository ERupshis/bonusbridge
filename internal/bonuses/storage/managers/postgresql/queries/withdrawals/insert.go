package withdrawals

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/erupshis/bonusbridge/internal/bonuses/data"
	dbBonusesData "github.com/erupshis/bonusbridge/internal/bonuses/storage/managers/postgresql/data"
	"github.com/erupshis/bonusbridge/internal/dberrors"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/retryer"
)

// Insert performs direct query request to database to add new withdrawal record.
func Insert(ctx context.Context, tx *sql.Tx, withdrawal *data.Withdrawal, log logger.BaseLogger) error {
	errMsg := fmt.Sprintf("insert withdrawal '%f' for userID '%d' in '%s'",
		withdrawal.Sum,
		withdrawal.UserID,
		dbBonusesData.GetTableFullName(dbBonusesData.WithdrawalsTable),
	) + ": %w"

	stmt, err := createInsertWithdrawalStmt(ctx, tx)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, log)

	query := func(context context.Context) error {
		_, err := stmt.ExecContext(
			context,
			withdrawal.UserID,
			withdrawal.Order,
			withdrawal.Sum,
		)

		return err
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, log, []int{1, 1, 3}, dberrors.DatabaseErrorsToRetry, query)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	return nil
}

// createInsertWithdrawalStmt generates statement for insert query.
func createInsertWithdrawalStmt(ctx context.Context, tx *sql.Tx) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	psqlInsert, _, err := psql.Insert(dbBonusesData.GetTableFullName(dbBonusesData.WithdrawalsTable)).
		Columns(dbBonusesData.ColumnsInWithdrawalsTable...).
		Values(make([]interface{}, len(dbBonusesData.ColumnsInWithdrawalsTable))...).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql insert statement for '%s': %w", dbBonusesData.GetTableFullName(dbBonusesData.WithdrawalsTable), err)
	}
	return tx.PrepareContext(ctx, psqlInsert)
}
