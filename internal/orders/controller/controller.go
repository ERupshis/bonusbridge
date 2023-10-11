package controller

import (
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/controller/handlers"
	"github.com/erupshis/bonusbridge/internal/orders/storage"
	"github.com/go-chi/chi/v5"
)

const packageName = "orders"

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
	r.Post("/", handlers.AddOrderHandler(c.storage, c.log))
	r.Get("/", handlers.GetOrdersHandler(c.storage, c.log))
	return r
}
