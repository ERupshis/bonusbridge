package bonuses

import (
	"github.com/erupshis/bonusbridge/internal/bonuses/handlers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/go-chi/chi/v5"
)

type Controller struct {
	log logger.BaseLogger
}

func CreateController(baseLogger logger.BaseLogger) Controller {
	return Controller{
		log: baseLogger,
	}
}

func (c *Controller) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", handlers.Balance(c.log))
	r.Post("/withdraw/", handlers.Withdraw(c.log))

	return r
}
