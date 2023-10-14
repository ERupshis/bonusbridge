package managers

import (
	"context"

	"github.com/erupshis/bonusbridge/internal/auth/users/data"
)

//go:generate mockgen -destination=../../../../mocks/mock_BaseUsersManager.go -package=mocks github.com/erupshis/bonusbridge/internal/auth/users/managers BaseUsersManager
type BaseUsersManager interface {
	AddUser(ctx context.Context, user *data.User) (int64, error)
	GetUser(ctx context.Context, login string) (*data.User, error)
	GetUserID(ctx context.Context, login string) (int64, error)
	GetUserRole(ctx context.Context, userID int64) (int, error)
}
