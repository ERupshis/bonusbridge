package queries

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	dbBonusesData "github.com/erupshis/bonusbridge/internal/bonuses/storage/managers/postgresql/data"
	"github.com/erupshis/bonusbridge/internal/db"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
	dbOrdersData "github.com/erupshis/bonusbridge/internal/orders/storage/managers/postgresql/data"
	"github.com/erupshis/bonusbridge/internal/retryer"
)

// Select performs direct query request to database to select orders satisfying filters.
func Select(ctx context.Context, tx *sql.Tx, filters map[string]interface{}, log logger.BaseLogger) ([]data.Order, error) {
	errMsg := fmt.Sprintf("select orders with filter '%v' in '%s'", filters, dbOrdersData.OrdersTable) + ": %w"

	stmt, err := createSelectOrdersStmt(ctx, tx, filters)
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
		dbOrdersData.GetTableFullName(dbOrdersData.StatusesTable),
		dbOrdersData.GetTableFullName(dbOrdersData.OrdersTable),
	)
	bonusesJoin := fmt.Sprintf("RIGHT JOIN %s ON %[1]s.id = %s.bonus_id",
		dbBonusesData.GetTableFullName(dbBonusesData.BonusesTable),
		dbOrdersData.GetTableFullName(dbOrdersData.OrdersTable),
	)

	builder := psql.Select(
		dbOrdersData.GetTableFullName(dbOrdersData.OrdersTable)+".id",
		dbOrdersData.GetTableFullName(dbOrdersData.OrdersTable)+".num",
		dbOrdersData.GetTableFullName(dbOrdersData.OrdersTable)+".user_id",
		dbOrdersData.GetTableFullName(dbOrdersData.StatusesTable)+".status",
		dbOrdersData.GetTableFullName(dbOrdersData.OrdersTable)+".bonus_id",
		dbBonusesData.GetTableFullName(dbBonusesData.BonusesTable)+".count",
		dbOrdersData.GetTableFullName(dbOrdersData.OrdersTable)+".uploaded_at",
	).
		From(dbOrdersData.GetTableFullName(dbOrdersData.OrdersTable)).
		JoinClause(statusesJoin).
		JoinClause(bonusesJoin)

	if len(filters) != 0 {
		for key := range filters {
			switch key {
			case "id":
				key = dbOrdersData.GetTableFullName(dbOrdersData.OrdersTable) + ".id"
			case "number":
				key = dbOrdersData.GetTableFullName(dbOrdersData.OrdersTable) + ".num"
			case "user_id":
				key = dbOrdersData.GetTableFullName(dbOrdersData.OrdersTable) + ".user_id"
			case "status_id":
				key = dbOrdersData.GetTableFullName(dbOrdersData.OrdersTable) + ".status_id"
			case "status":
				key = dbOrdersData.GetTableFullName(dbOrdersData.StatusesTable) + ".status"
			case "accrual":
				key = dbBonusesData.GetTableFullName(dbBonusesData.BonusesTable) + ".count"
			case "uploaded_at":
				key = dbOrdersData.GetTableFullName(dbOrdersData.OrdersTable) + ".uploaded_at"
			}
			builder = builder.Where(sq.Eq{key: "?"})
		}
	}
	psqlSelect, _, err := builder.ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql select statement for '%s': %w", dbOrdersData.GetTableFullName(dbOrdersData.OrdersTable), err)
	}
	return tx.PrepareContext(ctx, psqlSelect)
}
