package data

const (
	SchemaName = "users"
	UsersTable = "users"
)

// ColumnsInUsersTable slice of main table attributes in database.
var ColumnsInUsersTable = []string{"login", "password", "role_id"}

// GetTableFullName support function to return extended database table name.
func GetTableFullName(table string) string {
	return SchemaName + "." + table
}
