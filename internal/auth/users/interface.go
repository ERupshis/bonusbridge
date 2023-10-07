package users

const (
	RoleUser = iota
	RoleAdmin
)

type BaseUsers interface {
	AddUser(login string, password string) (int, error)
	GetUserId(login string) (int, error)
	GetUserRole(userID int) (int, error)
	ValidateUser(login string, password string) (bool, error)
}
