package controllers

import (
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/go-chi/chi/v5"
)

type controller struct {
	log logger.BaseLogger
}

func Create(baseLogger logger.BaseLogger) BaseController {
	return &controller{
		log: baseLogger,
	}
}

func (c *controller) Route() *chi.Mux {
	r := chi.NewRouter()

	return r
}
