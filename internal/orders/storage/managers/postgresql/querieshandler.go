package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
	"github.com/erupshis/bonusbridge/internal/retryer"
	"github.com/jackc/pgerrcode"
	"github.com/pkg/errors"
)

const (
	SchemaName    = "orders"
	OrdersTable   = "orders"
	StatusesTable = "statuses"
)

// columnsInOrdersTable slice of main table attributes in database.
var columnsInOrdersTable = []string{"num", "status_id", "user_id", "accrual_status", "uploaded_at"}

// databaseErrorsToRetry errors to retry request to database.
var databaseErrorsToRetry = []error{
	errors.New(pgerrcode.UniqueViolation),
	errors.New(pgerrcode.ConnectionException),
	errors.New(pgerrcode.ConnectionDoesNotExist),
	errors.New(pgerrcode.ConnectionFailure),
	errors.New(pgerrcode.SQLClientUnableToEstablishSQLConnection),
	errors.New(pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection),
	errors.New(pgerrcode.TransactionResolutionUnknown),
	errors.New(pgerrcode.ProtocolViolation),
}

// getTableFullName support function to return extended database table name.
func getTableFullName(table string) string {
	return SchemaName + "." + table
}

// QueriesHandler support object that implements database's queries and responsible for connection to databse.
type QueriesHandler struct {
	log logger.BaseLogger
}

// CreateHandler creates QueriesHandler.
func CreateHandler(log logger.BaseLogger) QueriesHandler {
	return QueriesHandler{log: log}
}

// InsertOrder performs direct query request to database to add new order.
func (q *QueriesHandler) InsertOrder(ctx context.Context, tx *sql.Tx, orderData *data.Order) (int64, error) {
	errorMsg := fmt.Sprintf("insert order '%v' in '%s'", *orderData, OrdersTable) + ": %w"

	stmt, err := createInsertOrderStmt(ctx, tx)
	if err != nil {
		return -1, fmt.Errorf(errorMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, q.log)

	newPersonId := int64(0)
	query := func(context context.Context) error {
		_, err := stmt.ExecContext(
			context,
			orderData.Number,
			data.GetOrderStatusID(orderData.Status),
			orderData.UserID,
			0,
			orderData.UploadedAt,
		)

		return err
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, q.log, []int{1, 1, 3}, databaseErrorsToRetry, query)
	if err != nil {
		return -1, fmt.Errorf(errorMsg, err)
	}

	return newPersonId, nil
}

// createInsertPersonStmt generates statement for insert query.
func createInsertOrderStmt(ctx context.Context, tx *sql.Tx) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	psqlInsert, _, err := psql.Insert(getTableFullName(OrdersTable)).
		Columns(columnsInOrdersTable...).
		Values(make([]interface{}, len(columnsInOrdersTable))...).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql insert statement for '%s': %w", getTableFullName(OrdersTable), err)
	}
	return tx.PrepareContext(ctx, psqlInsert)
}

// SelectOrders performs direct query request to database to select orders satisfying filters.
func (q *QueriesHandler) SelectOrders(ctx context.Context, tx *sql.Tx, filters map[string]interface{}) ([]data.Order, error) {
	errorMsg := fmt.Sprintf("select orders with filter '%v' in '%s'", filters, OrdersTable) + ": %w"

	stmt, err := createSelectOrdersStmt(ctx, tx, filters)
	if err != nil {
		return nil, fmt.Errorf(errorMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, q.log)

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
		return err
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, q.log, []int{1, 1, 3}, databaseErrorsToRetry, query)
	if err != nil {
		return nil, fmt.Errorf(errorMsg, err)
	}

	defer helpers.ExecuteWithLogError(rows.Close, q.log)
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

	statusesJoin := fmt.Sprintf("LEFT JOIN %s ON %[1]s.id = %s.status_id", getTableFullName(StatusesTable), getTableFullName(OrdersTable))
	builder := psql.Select(
		getTableFullName(OrdersTable)+".id",
		"num",
		"user_id",
		"status",
		"accrual_status",
		"uploaded_at",
	).
		From(getTableFullName(OrdersTable)).
		JoinClause(statusesJoin)
	if len(filters) != 0 {
		for key := range filters {
			switch key {
			case "id":
				key = getTableFullName(OrdersTable) + ".id"
			case "number":
				key = getTableFullName(OrdersTable) + ".num"
			case "user_id":
				key = getTableFullName(OrdersTable) + ".user_id"
			case "status":
				key = getTableFullName(StatusesTable) + ".status"
			case "accrual":
				key = getTableFullName(OrdersTable) + ".accrual_status"
			case "uploaded_at":
				key = getTableFullName(OrdersTable) + ".uploaded_at"
			}
			builder = builder.Where(sq.Eq{key: "?"})
		}
	}
	psqlSelect, _, err := builder.ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql select statement for '%s': %w", getTableFullName(OrdersTable), err)
	}
	return tx.PrepareContext(ctx, psqlSelect)
}

/*
// DeletePerson performs direct query request to database to delete person by id.
func (q *QueriesHandler) DeletePerson(ctx context.Context, tx *sql.Tx, id int64) (int64, error) {
	errorMsg := fmt.Sprintf("delete person by id '%v' in '%s", id, PersonsTable) + ": %w"

	stmt, err := createDeletePersonStmt(ctx, tx)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, q.log)

	var result sql.Result
	query := func(context context.Context) error {
		result, err = stmt.ExecContext(context, id)
		return err
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, q.log, []int{1, 1, 3}, databaseErrorsToRetry, query)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	return count, nil
}

// createDeletePersonStmt generates statement for delete query.
func createDeletePersonStmt(ctx context.Context, tx *sql.Tx) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	psqlInsert, _, err := psql.Delete(getTableFullName(PersonsTable)).
		Where(sq.Eq{"id": "?"}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql delete statement for '"+getTableFullName(PersonsTable)+"': %w", err)

	}
	return tx.PrepareContext(ctx, psqlInsert)
}

// UpdatePartialPersonById generates statement for delete query.
func (q *QueriesHandler) UpdatePartialPersonById(ctx context.Context, tx *sql.Tx, id int64, values map[string]interface{}) (int64, error) {
	errorMsg := fmt.Sprintf("update partially person by id '%d' with data '%v' in '%s'", id, values, PersonsTable) + ": %w"

	var columnsToUpdate []string
	var valuesToUpdate []interface{}
	for key, val := range values {
		columnsToUpdate = append(columnsToUpdate, key)
		valuesToUpdate = append(valuesToUpdate, val)
	}
	valuesToUpdate = append(valuesToUpdate, id)

	stmt, err := createUpdatePersonByIdStmt(ctx, tx, columnsToUpdate)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, q.log)

	var result sql.Result
	query := func(context context.Context) error {
		result, err = stmt.ExecContext(
			context,
			valuesToUpdate...,
		)
		return err
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, q.log, []int{1, 1, 3}, databaseErrorsToRetry, query)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	return count, nil
}

// createUpdatePersonByIdStmt generates statement for update query.
func createUpdatePersonByIdStmt(ctx context.Context, tx *sql.Tx, values []string) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.Update(getTableFullName(PersonsTable))
	for _, col := range values {
		builder = builder.Set(col, "?")
	}
	builder = builder.Where(sq.Eq{"id": "?"})
	psqlUpdate, _, err := builder.ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql update statement for '"+getTableFullName(PersonsTable)+"': %w", err)

	}
	return tx.PrepareContext(ctx, psqlUpdate)
}

// GetAdditionalId returns foreign key from linked table.
func (q *QueriesHandler) GetAdditionalId(ctx context.Context, tx *sql.Tx, name string, table string) (int64, error) {
	errorMsg := fmt.Sprintf("get additional id for '%s' in '%s'", name, table) + ": %w"

	stmt, err := createSelectAdditionalIdStmt(ctx, tx, name, table)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, q.log)

	var id int64
	query := func(context context.Context) error {
		return stmt.QueryRowContext(ctx, name).Scan(&id)
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, q.log, []int{1, 1, 3}, databaseErrorsToRetry, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			id, err = q.InsertAdditionalId(ctx, tx, name, table)
			if err != nil {
				return 0, fmt.Errorf(errorMsg, err)
			}
		} else {
			return 0, fmt.Errorf(errorMsg, err)
		}
	}

	return id, nil
}

// createSelectAdditionalIdStmt generates statement for get foreign key id.
func createSelectAdditionalIdStmt(ctx context.Context, tx *sql.Tx, name string, table string) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	psqlSelect, _, err := psql.Select("id").
		From(getTableFullName(table)).
		Where(sq.Eq{"name": name}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql select statement for '"+getTableFullName(table)+"': %w", err)

	}
	return tx.PrepareContext(ctx, psqlSelect)
}

// InsertAdditionalId adds new value forl linked table and returns foreign key from linked table.
func (q *QueriesHandler) InsertAdditionalId(ctx context.Context, tx *sql.Tx, name string, table string) (int64, error) {
	errorMsg := fmt.Sprintf("insert additional value for '%s' in '%s'", name, table) + ": %w"

	stmt, err := createInsertAdditionalIdStmt(ctx, tx, name, table)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}
	defer helpers.ExecuteWithLogError(stmt.Close, q.log)

	var id int64
	query := func(context context.Context) error {
		return stmt.QueryRowContext(ctx, name).Scan(&id)
	}
	err = retryer.RetryCallWithTimeoutErrorOnly(ctx, q.log, []int{1, 1, 3}, databaseErrorsToRetry, query)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	return id, nil
}

// createInsertAdditionalIdStmt generates statement for add and then get foreign key id.
func createInsertAdditionalIdStmt(ctx context.Context, tx *sql.Tx, name string, table string) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	psqlInsert, _, err := psql.Insert(getTableFullName(table)).
		Columns("name").
		Values(name).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel sql insert statement for '"+getTableFullName(table)+"': %w", err)

	}
	return tx.PrepareContext(ctx, psqlInsert+"RETURNING id")
}
*/
