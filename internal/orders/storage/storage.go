package storage

import (
	"context"
	"fmt"

	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/storage/data"
	"github.com/erupshis/bonusbridge/internal/orders/storage/managers"
)

var ErrOrderWasAddedByAnotherUser = fmt.Errorf("order has already been added by another user")
var ErrOrderWasAddedBefore = fmt.Errorf("order has already been added")

type Storage struct {
	manager managers.BaseStorageManager

	log logger.BaseLogger
}

func Create(manager managers.BaseStorageManager, baseLogger logger.BaseLogger) Storage {
	return Storage{
		manager: manager,
		log:     baseLogger,
	}
}

func (s *Storage) AddOrder(ctx context.Context, number string, userID int64) error {
	order, err := s.manager.GetOrder(number)
	if err != nil {
		return fmt.Errorf("get order from storage: %w", err)
	}

	if order == nil {
		_, err = s.manager.AddOrder(ctx, number, userID)
		if err != nil {
			return fmt.Errorf("add new order in storage: %w", err)
		}

		return nil
	} else {
		if order.UserID == userID {
			return fmt.Errorf("add order in storage: %w", ErrOrderWasAddedBefore)
		} else {
			return fmt.Errorf("add order in storage: %w", ErrOrderWasAddedByAnotherUser)
		}
	}
}

func (s *Storage) GetOrders(userID int64) ([]data.Order, error) {
	orders, err := s.manager.GetOrders(userID)
	if err != nil {
		return nil, fmt.Errorf("get orders from storage: %w", err)
	}

	return orders, nil
}
