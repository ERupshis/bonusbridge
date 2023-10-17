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

func (s *Storage) WithdrawBonuses(ctx context.Context, withdrawal *data.Withdrawal) error {
	balanceDif, err := s.manager.GetBalanceDif(ctx, withdrawal.UserID)
	if err != nil {
		return fmt.Errorf("get userID '%d' bonuses balance: %w", withdrawal.UserID, err)
	}

	if balanceDif < withdrawal.Sum {
		return fmt.Errorf("userID '%d' balance '%f' is not enough for withdrawn: %w", withdrawal.UserID, balanceDif, data.ErrNotEnoughBonuses)
	}

	if err = s.manager.WithdrawBonuses(ctx, withdrawal); err != nil {
		return fmt.Errorf("withdraw userID '%d' bonuses: %w", withdrawal.UserID, err)
	}

	return nil
}

func (s *Storage) GetBalance(ctx context.Context, userID int64) (*data.Balance, error) {
	var res data.Balance

	var err error
	res.Current, err = s.manager.GetBalance(ctx, true, userID)
	if err != nil {
		return nil, fmt.Errorf("get userID '%d' bonuses balance: %w", userID, err)
	}

	res.Withdrawn, err = s.manager.GetBalance(ctx, false, userID)
	if err != nil {
		return nil, fmt.Errorf("get userID '%d' bonuses balance: %w", userID, err)
	}

	if res.Withdrawn < 0 {
		res.Withdrawn = -res.Withdrawn
	}

	return &res, nil
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
