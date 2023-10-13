package managers

import (
	"context"

	"github.com/erupshis/bonusbridge/internal/bonuses/data"
)

type BaseBonusesManager interface {
	AddBonuses(ctx context.Context, userID int64, count float32) error
	GetBalance(ctx context.Context, userID int64) (*data.Balance, error)
	WithdrawBonuses(ctx context.Context, withdrawal *data.Withdrawal) error
	GetWithdrawals(ctx context.Context, userID int64) ([]data.Withdrawal, error)
}
