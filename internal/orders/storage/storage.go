package storage

import (
	"context"
	"fmt"

	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
	"github.com/erupshis/bonusbridge/internal/orders/storage/managers"
)

type Storage struct {
	manager managers.BaseOrdersManager

	log logger.BaseLogger
}

func Create(manager managers.BaseOrdersManager, baseLogger logger.BaseLogger) Storage {
	return Storage{
		manager: manager,
		log:     baseLogger,
	}
}

func (s *Storage) AddOrder(ctx context.Context, number string, userID int64) error {
	order, err := s.manager.GetOrder(ctx, number)
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
			return fmt.Errorf("add order in storage: %w", data.ErrOrderWasAddedBefore)
		} else {
			return fmt.Errorf("add order in storage: %w", data.ErrOrderWasAddedByAnotherUser)
		}
	}
}

func (s *Storage) UpdateOrder(ctx context.Context, order *data.Order) error {
	if err := s.manager.UpdateOrder(ctx, order); err != nil {
		return fmt.Errorf("update order in storage: %w", err)
	}

	return nil
}

func (s *Storage) GetOrders(ctx context.Context, filters map[string]interface{}) ([]data.Order, error) {
	orders, err := s.manager.GetOrders(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("get orders from storage: %w", err)
	}

	return orders, nil
}
