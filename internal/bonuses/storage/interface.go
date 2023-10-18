package storage

import (
	"context"

	"github.com/erupshis/bonusbridge/internal/bonuses/data"
)

//go:generate mockgen -destination=../../../mocks/mock_BaseBonusesStorage.go -package=mocks github.com/erupshis/bonusbridge/internal/bonuses/storage BaseBonusesStorage
type BaseBonusesStorage interface {
	WithdrawBonuses(ctx context.Context, withdrawal *data.Withdrawal) error
	GetBalance(ctx context.Context, userID int64) (*data.Balance, error)
	GetWithdrawals(ctx context.Context, userID int64) ([]data.Withdrawal, error)
}
