package bonuses

const (
	BonusesTable     = "bonuses"
	WithdrawalsTable = "withdrawals"
)

// ColumnsInBonusesTable slice of main table attributes in database.
var ColumnsInBonusesTable = []string{"user_id", "count"}

// ColumnsInWithdrawalsTable slice of main table attributes in database.
var ColumnsInWithdrawalsTable = []string{"user_id", "order_num", "bonus_id", "processed_at"}
