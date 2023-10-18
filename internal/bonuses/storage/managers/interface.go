package managers

import (
	"context"

	"github.com/erupshis/bonusbridge/internal/bonuses/data"
)

//go:generate mockgen -destination=../../../../mocks/mock_BaseBonusesManager.go -package=mocks github.com/erupshis/bonusbridge/internal/bonuses/storage/managers BaseBonusesManager
type BaseBonusesManager interface {
	GetBalanceDif(ctx context.Context, userID int64) (float32, error)
	GetBalance(ctx context.Context, income bool, userID int64) (float32, error)

	WithdrawBonuses(ctx context.Context, withdrawal *data.Withdrawal) error
	GetWithdrawals(ctx context.Context, userID int64) ([]data.Withdrawal, error)
}
