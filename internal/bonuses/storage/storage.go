package storage

import (
	"context"

	"github.com/erupshis/bonusbridge/internal/bonuses/data"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage/managers"
	"github.com/erupshis/bonusbridge/internal/logger"
)

type Storage struct {
	manager managers.BaseBonusesManager

	log logger.BaseLogger
}

func Create(manager managers.BaseBonusesManager, baseLogger logger.BaseLogger) Storage {
	return Storage{
		manager: manager,
		log:     baseLogger,
	}
}

func (s *Storage) AddBonuses(ctx context.Context, userID int64, count int64) error {
	return nil
}

func (s *Storage) WithdrawBonuses(ctx context.Context, userID int64, count int64) error {
	return nil
}

func (s *Storage) GetBalance(ctx context.Context, userID int64) (*data.Balance, error) {
	return nil, nil
}

func (s *Storage) GetWithdrawals(ctx context.Context, userID int64) ([]data.Withdrawal, error) {
	return nil, nil
}
