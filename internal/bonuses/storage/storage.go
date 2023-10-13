package storage

import (
	"context"
	"fmt"

	"github.com/erupshis/bonusbridge/internal/bonuses/data"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage/managers"
	"github.com/erupshis/bonusbridge/internal/logger"
)

var ErrNotEnoughBonuses = fmt.Errorf("not enough bonuses for withdrawal")

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

func (s *Storage) WithdrawBonuses(ctx context.Context, userID int64, count float32) error {
	balance, err := s.manager.GetBonuses(ctx, userID)
	if err != nil {
		return fmt.Errorf("get userID '%d' bonuses balance: %w", userID, err)
	}

	if balance < count {
		return fmt.Errorf("userID '%d' balance '%f' is not enough for withdrawn: %w", userID, balance, ErrNotEnoughBonuses)
	}

	if err = s.manager.WithdrawBonuses(ctx, userID, count); err != nil {
		return fmt.Errorf("withdraw userID '%d' bonuses: %w", userID, err)
	}

	return nil
}

func (s *Storage) GetBalance(ctx context.Context, userID int64) (*data.Balance, error) {
	var balance data.Balance

	bonuses, err := s.manager.GetBonuses(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get userID '%d' bonuses balance: %w", userID, err)
	}

	balance.Current = bonuses

	withdrawn, err := s.manager.GetWithdrawnBonuses(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get userID '%d' withdrawn bonuses: %w", userID, err)
	}

	balance.Withdrawn = withdrawn

	return &balance, nil
}

func (s *Storage) GetWithdrawals(ctx context.Context, userID int64) ([]data.Withdrawal, error) {
	return s.manager.GetWithdrawals(ctx, userID)
}
