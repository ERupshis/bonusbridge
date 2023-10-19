package orders

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/erupshis/bonusbridge/internal/db"
	"github.com/erupshis/bonusbridge/internal/db/queries"
	dbBonusesData "github.com/erupshis/bonusbridge/internal/db/queries/bonuses"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
	"github.com/erupshis/bonusbridge/internal/retryer"
)

// Select performs direct query request to database to select orders satisfying filters.
func Select(ctx context.Context, tx *sql.Tx, filters map[string]interface{}, log logger.BaseLogger) ([]data.Order, error) {
	errMsg := fmt.Sprintf("select orders with filter '%v' in '%s'", filters, OrdersTable) + ": %w"

	stmt, err := createSelectOrdersStmt(ctx, tx, filters)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, log)

	var valuesToUpdate []interface{}
	for key, val := range filters {
		if key == queries.Custom {
			continue
		}

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
	var res []data.Order
	for rows.Next() {
		order := data.Order{}
		err := rows.Scan(
			&order.ID,
			&order.Number,
			&order.UserID,
			&order.Status,
			&order.BonusID,
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

	statusesJoin := fmt.Sprintf("RIGHT JOIN %s ON %[1]s.id = %s.status_id",
		GetTableFullName(StatusesTable),
		GetTableFullName(OrdersTable),
	)
	bonusesJoin := fmt.Sprintf("RIGHT JOIN %s ON %[1]s.id = %s.bonus_id",
		dbBonusesData.GetTableFullName(dbBonusesData.BonusesTable),
		GetTableFullName(OrdersTable),
	)

	builder := psql.Select(
		GetTableFullName(OrdersTable)+".id",
		GetTableFullName(OrdersTable)+".num",
		GetTableFullName(OrdersTable)+".user_id",
		GetTableFullName(StatusesTable)+".status",
		GetTableFullName(OrdersTable)+".bonus_id",
		dbBonusesData.GetTableFullName(dbBonusesData.BonusesTable)+".count",
		GetTableFullName(OrdersTable)+".uploaded_at",
	).
		From(GetTableFullName(OrdersTable)).
		JoinClause(statusesJoin).
		JoinClause(bonusesJoin)

	if len(filters) != 0 {
		for key, val := range filters {
			switch key {
			case "id":
				key = GetTableFullName(OrdersTable) + ".id"
			case "number":
				key = GetTableFullName(OrdersTable) + ".num"
			case "user_id":
				key = GetTableFullName(OrdersTable) + ".user_id"
			case "status_id":
				key = GetTableFullName(OrdersTable) + ".status_id"
			case "status":
				key = GetTableFullName(StatusesTable) + ".status"
			case "accrual":
				key = dbBonusesData.GetTableFullName(dbBonusesData.BonusesTable) + ".count"
			case "uploaded_at":
				key = GetTableFullName(OrdersTable) + ".uploaded_at"
			case queries.Custom:
				builder = builder.Where(val)
				continue
			}
			builder = builder.Where(sq.Eq{key: "?"})
		}
	}
	psqlSelect, _, err := builder.ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql select statement for '%s': %w", GetTableFullName(OrdersTable), err)
	}
	return tx.PrepareContext(ctx, psqlSelect)
}
