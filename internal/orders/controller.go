package orders

import (
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/handlers"
	"github.com/erupshis/bonusbridge/internal/orders/storage"
	"github.com/go-chi/chi/v5"
)

type Controller struct {
	storage storage.BaseOrdersStorage

	log logger.BaseLogger
}

func CreateController(storage storage.BaseOrdersStorage, baseLogger logger.BaseLogger) Controller {
	return Controller{
		storage: storage,
		log:     baseLogger,
	}
}

func (c *Controller) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", handlers.AddOrder(c.storage, c.log))
	r.Get("/", handlers.GetOrders(c.storage, c.log))
	return r
}
