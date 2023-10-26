package orders

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/erupshis/bonusbridge/internal/db"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
	"github.com/erupshis/bonusbridge/internal/retryer"
)

// Insert performs direct query request to database to add new order.
func Insert(ctx context.Context, tx *sql.Tx, orderData *data.Order, log logger.BaseLogger) (int64, error) {
	errMsg := fmt.Sprintf("insert order '%v' in '%s'", *orderData, OrdersTable) + ": %w"

	stmt, err := createInsertOrderStmt(ctx, tx)
	if err != nil {
		return -1, fmt.Errorf(errMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, log)

	newOrderID := int64(0)
	query := func(context context.Context) error {
		err := stmt.QueryRowContext(
			context,
			orderData.Number,
			data.GetOrderStatusID(orderData.Status),
			orderData.UserID,
			orderData.BonusID,
			orderData.UploadedAt,
		).Scan(&newOrderID)

		return err
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, log, []int{1, 1, 3}, db.DatabaseErrorsToRetry, query)
	if err != nil {
		return -1, fmt.Errorf(errMsg, err)
	}

	return newOrderID, nil
}

// createInsertPersonStmt generates statement for insert query.
func createInsertOrderStmt(ctx context.Context, tx *sql.Tx) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	psqlInsert, _, err := psql.Insert(OrdersTable).
		Columns(ColumnsInOrdersTable...).
		Values(make([]interface{}, len(ColumnsInOrdersTable))...).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql insert statement for '%s': %w", OrdersTable, err)
	}
	return tx.PrepareContext(ctx, psqlInsert)
}
