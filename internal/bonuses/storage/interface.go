package storage

import (
	"context"

	"github.com/erupshis/bonusbridge/internal/bonuses/data"
)

type BaseBonusesStorage interface {
	WithdrawBonuses(ctx context.Context, withdrawal *data.Withdrawal) error
	GetBalance(ctx context.Context, userID int64) (*data.Balance, error)
	GetWithdrawals(ctx context.Context, userID int64) ([]data.Withdrawal, error)
}
