package users

const (
	SchemaName = "users"
	UsersTable = "users"
)

// ColumnsInUsersTable slice of main table attributes in database.
var ColumnsInUsersTable = []string{"login", "password", "role_id"}
