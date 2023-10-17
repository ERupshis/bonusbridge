package data

const (
	SchemaName       = "bonuses"
	BonusesTable     = "bonuses"
	WithdrawalsTable = "withdrawals"
)

// ColumnsInBonusesTable slice of main table attributes in database.
var ColumnsInBonusesTable = []string{"user_id", "count"}

// ColumnsInWithdrawalsTable slice of main table attributes in database.
var ColumnsInWithdrawalsTable = []string{"user_id", "order_id", "bonus_id", "processed_at"}

// GetTableFullName support function to return extended database table name.
func GetTableFullName(table string) string {
	return SchemaName + "." + table
}
