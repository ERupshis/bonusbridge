package data

const (
	SchemaName       = "bonuses"
	BonusesTable     = "bonuses"
	WithdrawalsTable = "withdrawals"
)

// ColumnsInBonusesTable slice of main table attributes in database.
var ColumnsInBonusesTable = []string{"user_id", "balance", "withdrawn"}

// ColumnsInWithdrawalsTable slice of main table attributes in database.
var ColumnsInWithdrawalsTable = []string{"user_id", "order_num", "sum", "processed_at"}

// GetTableFullName support function to return extended database table name.
func GetTableFullName(table string) string {
	return SchemaName + "." + table
}
