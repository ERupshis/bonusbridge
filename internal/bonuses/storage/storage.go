package storage

import (
	"context"
	"fmt"

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

func (s *Storage) AddBonuses(ctx context.Context, userID int64, count float32) error {
	return s.manager.AddBonuses(ctx, userID, count)
}

func (s *Storage) WithdrawBonuses(ctx context.Context, withdrawal *data.Withdrawal) error {
	balance, err := s.manager.GetBalance(ctx, withdrawal.UserID)
	if err != nil {
		return fmt.Errorf("get userID '%d' bonuses balance: %w", withdrawal.UserID, err)
	}

	//TODO: to remove
	if balance == nil {
		if err = s.manager.AddBonuses(ctx, withdrawal.UserID, 0); err != nil {
			return fmt.Errorf("init bonuses for userID '%d': %w", withdrawal.UserID, err)
		}
		balance = &data.Balance{UserID: withdrawal.UserID}
	}

	if balance.Current < withdrawal.Sum {
		return fmt.Errorf("userID '%d' balance '%f' is not enough for withdrawn: %w", withdrawal.UserID, balance.Current, data.ErrNotEnoughBonuses)
	}

	if err = s.manager.WithdrawBonuses(ctx, withdrawal); err != nil {
		return fmt.Errorf("withdraw userID '%d' bonuses: %w", withdrawal.UserID, err)
	}

	return nil
}

func (s *Storage) GetBalance(ctx context.Context, userID int64) (*data.Balance, error) {
	balance, err := s.manager.GetBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get userID '%d' bonuses balance: %w", userID, err)
	}

	if balance == nil {
		if err = s.AddBonuses(ctx, userID, 0); err != nil {
			return nil, fmt.Errorf("init balance for userID '%d': %w", userID, err)
		}
		balance = &data.Balance{UserID: userID}
	}

	return balance, nil
}

func (s *Storage) GetWithdrawals(ctx context.Context, userID int64) ([]data.Withdrawal, error) {
	withdrawals, err := s.manager.GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get userID '%d' withdrawals: %w", userID, err)
	}

	if len(withdrawals) == 0 {
		return nil, fmt.Errorf("get userID '%d' withdrawals: %w", userID, data.ErrWithdrawalsMissing)
	}

	return withdrawals, nil
}
