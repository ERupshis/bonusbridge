package bonuses

import (
	"github.com/erupshis/bonusbridge/internal/bonuses/handlers"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage"
	"github.com/erupshis/bonusbridge/internal/logger"
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
	r.Get("/", handlers.Balance(c.storage, c.log))
	r.Post("/withdraw", handlers.Withdraw(c.storage, c.log))

	return r
}
