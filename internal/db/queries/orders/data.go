package orders

const (
	OrdersTable   = "orders"
	StatusesTable = "statuses"
)

// ColumnsInOrdersTable slice of main table attributes in database.
var ColumnsInOrdersTable = []string{"num", "status_id", "user_id", "bonus_id", "uploaded_at"}
