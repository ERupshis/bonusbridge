package managers

import (
	"context"

	"github.com/erupshis/bonusbridge/internal/bonuses/data"
)

type BaseBonusesManager interface {
	AddBonuses(ctx context.Context, userID int64, count int64) error
	GetBonuses(ctx context.Context, userID int64) (int64, error)
	WithdrawBonuses(ctx context.Context, userID int64, count int64) (bool, error)
	GetWithdrawnBonuses(ctx context.Context, userID int64) (int64, error)
	GetWithdrawals(ctx context.Context, userID int64) ([]data.Withdrawal, error)
}
