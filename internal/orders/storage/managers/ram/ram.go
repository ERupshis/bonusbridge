package ram

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/storage/data"
	"github.com/erupshis/bonusbridge/internal/orders/storage/managers"
)

type manager struct {
	mu     sync.RWMutex
	orders []data.Order

	log logger.BaseLogger
}

func Create(baseLogger logger.BaseLogger) managers.BaseStorageManager {
	baseLogger.Info("[ram:Create] ram storage created")
	return &manager{
		log:    baseLogger,
		orders: make([]data.Order, 0),
	}
}

func (m *manager) AddOrder(_ context.Context, number string, userID int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orders = append(m.orders, data.Order{
		ID:         len(m.orders),
		Number:     number,
		UserID:     userID,
		Status:     "NEW",
		Accrual:    "",
		UploadedAt: time.Now(),
	})
	return int64(len(m.orders)), nil
}

func (m *manager) GetOrder(number string) (*data.Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, order := range m.orders {
		if order.Number == number {
			return &order, nil
		}
	}
	return nil, nil
}

func (m *manager) GetOrders(userID int64) ([]data.Order, error) {
	m.mu.RLock()
	orders := make([]data.Order, 0)
	for _, order := range m.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	m.mu.RUnlock()

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].UploadedAt.Before(orders[j].UploadedAt)
	})

	return orders, nil
}
