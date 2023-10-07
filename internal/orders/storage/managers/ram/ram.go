package ram

import (
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/storage/data"
	"github.com/erupshis/bonusbridge/internal/orders/storage/managers"
)

type manager struct {
	orders []data.Order

	log logger.BaseLogger
}

func Create(baseLogger logger.BaseLogger) managers.BaseStorageManager {
	return &manager{
		log:    baseLogger,
		orders: make([]data.Order, 0),
	}
}

func (m *manager) AddOrder(number string, userID int64) error {
	return nil
}

func (m *manager) GetOrders(userID int64) ([]data.Order, error) {
	return nil, nil
}
