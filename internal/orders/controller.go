package orders

import (
	"github.com/erupshis/bonusbridge/internal/logger"
	handlers2 "github.com/erupshis/bonusbridge/internal/orders/handlers"
	"github.com/erupshis/bonusbridge/internal/orders/storage"
	"github.com/go-chi/chi/v5"
)

type Controller struct {
	storage storage.Storage

	log logger.BaseLogger
}

func CreateController(storage storage.Storage, baseLogger logger.BaseLogger) Controller {
	return Controller{
		storage: storage,
		log:     baseLogger,
	}
}

func (c *Controller) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", handlers2.AddOrderHandler(c.storage, c.log))
	r.Get("/", handlers2.GetOrdersHandler(c.storage, c.log))
	return r
}
