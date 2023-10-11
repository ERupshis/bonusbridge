package data

import (
	"errors"

	"github.com/jackc/pgerrcode"
)

const (
	SchemaName    = "orders"
	OrdersTable   = "orders"
	StatusesTable = "statuses"
)

// ColumnsInOrdersTable slice of main table attributes in database.
var ColumnsInOrdersTable = []string{"num", "status_id", "user_id", "accrual_status", "uploaded_at"}

// DatabaseErrorsToRetry errors to retry request to database.
var DatabaseErrorsToRetry = []error{
	errors.New(pgerrcode.UniqueViolation),
	errors.New(pgerrcode.ConnectionException),
	errors.New(pgerrcode.ConnectionDoesNotExist),
	errors.New(pgerrcode.ConnectionFailure),
	errors.New(pgerrcode.SQLClientUnableToEstablishSQLConnection),
	errors.New(pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection),
	errors.New(pgerrcode.TransactionResolutionUnknown),
	errors.New(pgerrcode.ProtocolViolation),
}

// GetTableFullName support function to return extended database table name.
func GetTableFullName(table string) string {
	return SchemaName + "." + table
}
