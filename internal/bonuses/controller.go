package bonuses

import (
	"github.com/erupshis/bonusbridge/internal/bonuses/handlers"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/go-chi/chi/v5"
)

type Controller struct {
	storage storage.BaseBonusesStorage

	log logger.BaseLogger
}

func CreateController(storage storage.BaseBonusesStorage, baseLogger logger.BaseLogger) Controller {
	return Controller{
		storage: storage,
		log:     baseLogger,
	}
}

func (c *Controller) RouteBonuses() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", handlers.Balance(c.storage, c.log))
	r.Post("/withdraw", handlers.Withdraw(c.storage, c.log))

	return r
}

func (c *Controller) RouteWithdrawals() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", handlers.Withdrawals(c.storage, c.log))

	return r
}
