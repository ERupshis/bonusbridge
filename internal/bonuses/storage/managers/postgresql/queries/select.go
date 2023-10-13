package queries

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
	dbData "github.com/erupshis/bonusbridge/internal/orders/storage/managers/postgresql/data"
	"github.com/erupshis/bonusbridge/internal/retryer"
)

// SelectOrders performs direct query request to database to select orders satisfying filters.
func SelectOrders(ctx context.Context, tx *sql.Tx, filters map[string]interface{}, log logger.BaseLogger) ([]data.Order, error) {
	errorMsg := fmt.Sprintf("select orders with filter '%v' in '%s'", filters, dbData.OrdersTable) + ": %w"

	stmt, err := createSelectOrdersStmt(ctx, tx, filters)
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
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, log, []int{1, 1, 3}, dbData.DatabaseErrorsToRetry, query)
	if err != nil {
		return nil, fmt.Errorf(errorMsg, err)
	}

	defer helpers.ExecuteWithLogError(rows.Close, log)
	var res []data.Order
	for rows.Next() {
		order := data.Order{}
		err := rows.Scan(
			&order.ID,
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("parse db result: %w", err)
		}

		res = append(res, order)
	}

	return res, nil
}

// createSelectOrdersStmt generates statement for select query.
func createSelectOrdersStmt(ctx context.Context, tx *sql.Tx, filters map[string]interface{}) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	statusesJoin := fmt.Sprintf("LEFT JOIN %s ON %[1]s.id = %s.status_id", dbData.GetTableFullName(dbData.StatusesTable), dbData.GetTableFullName(dbData.OrdersTable))
	builder := psql.Select(
		dbData.GetTableFullName(dbData.OrdersTable)+".id",
		"num",
		"user_id",
		"status",
		"accrual_status",
		"uploaded_at",
	).
		From(dbData.GetTableFullName(dbData.OrdersTable)).
		JoinClause(statusesJoin)
	if len(filters) != 0 {
		for key := range filters {
			switch key {
			case "id":
				key = dbData.GetTableFullName(dbData.OrdersTable) + ".id"
			case "number":
				key = dbData.GetTableFullName(dbData.OrdersTable) + ".num"
			case "user_id":
				key = dbData.GetTableFullName(dbData.OrdersTable) + ".user_id"
			case "status":
				key = dbData.GetTableFullName(dbData.StatusesTable) + ".status"
			case "accrual":
				key = dbData.GetTableFullName(dbData.OrdersTable) + ".accrual_status"
			case "uploaded_at":
				key = dbData.GetTableFullName(dbData.OrdersTable) + ".uploaded_at"
			}
			builder = builder.Where(sq.Eq{key: "?"})
		}
	}
	psqlSelect, _, err := builder.ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql select statement for '%s': %w", dbData.GetTableFullName(dbData.OrdersTable), err)
	}
	return tx.PrepareContext(ctx, psqlSelect)
}
