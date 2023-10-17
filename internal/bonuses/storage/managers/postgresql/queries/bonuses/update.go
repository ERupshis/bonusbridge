package bonuses

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage/managers/postgresql/data"
	"github.com/erupshis/bonusbridge/internal/db"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/retryer"
)

// UpdateByID performs direct query request to database to edit existing bonuses record.
func UpdateByID(ctx context.Context, tx *sql.Tx, id int64, values map[string]interface{}, log logger.BaseLogger) error {
	errMsg := fmt.Sprintf("update partially bonus by id '%d' with data '%v' in '%s'", id, values, data.GetTableFullName(data.BonusesTable)) + ": %w"

	var columnsToUpdate []string
	var valuesToUpdate []interface{}
	for key, val := range values {
		columnsToUpdate = append(columnsToUpdate, key)
		valuesToUpdate = append(valuesToUpdate, val)
	}
	valuesToUpdate = append(valuesToUpdate, id)

	stmt, err := createUpdateByIDStmt(ctx, tx, columnsToUpdate)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, log)

	var result sql.Result
	query := func(context context.Context) error {
		result, err = stmt.ExecContext(
			context,
			valuesToUpdate...,
		)
		return err
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, log, []int{1, 1, 3}, db.DatabaseErrorsToRetry, query)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	return nil
}

// createUpdateByIDStmt generates statement for update query.
func createUpdateByIDStmt(ctx context.Context, tx *sql.Tx, values []string) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.Update(data.GetTableFullName(data.BonusesTable))
	for _, col := range values {
		builder = builder.Set(col, "?")
	}
	builder = builder.Where(sq.Eq{"id": "?"})
	psqlUpdate, _, err := builder.ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql update statement for '%s': %w", data.GetTableFullName(data.BonusesTable), err)

	}
	return tx.PrepareContext(ctx, psqlUpdate)
}
