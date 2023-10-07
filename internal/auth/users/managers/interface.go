package managers

const (
	RoleUser = iota
	RoleAdmin
)

//go:generate mockgen -destination=../../../../mocks/mock_BaseUsersManager.go -package=mocks github.com/erupshis/bonusbridge/internal/auth/users/managers BaseUsersManager
type BaseUsersManager interface {
	AddUser(login string, password string) (int, error)
	GetUserId(login string) (int, error)
	GetUserRole(userID int) (int, error)
	ValidateUser(login string, password string) (bool, error)
}
