package storage

import (
	"fmt"

	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/storage/managers"
)

var ErrOrderWasAddedByAnotherUser = fmt.Errorf("order has already been added by another user")
var ErrOrderWasAddedBefore = fmt.Errorf("order has already been added before")

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

func (s *Storage) AddOrder(number string, userID int64) error {
	//TODO: check order presence, compare userID if exists. return error
	//TODO: if missing - add new order in system.

	return nil
}
