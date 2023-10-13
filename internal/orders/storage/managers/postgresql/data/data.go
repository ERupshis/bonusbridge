package data

const (
	SchemaName    = "orders"
	OrdersTable   = "orders"
	StatusesTable = "statuses"
)

// ColumnsInOrdersTable slice of main table attributes in database.
var ColumnsInOrdersTable = []string{"num", "status_id", "user_id", "accrual", "uploaded_at"}

// GetTableFullName support function to return extended database table name.
func GetTableFullName(table string) string {
	return SchemaName + "." + table
}
