package controllers

import (
	"net/http"

	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/go-chi/chi"
)

type controller struct {
	log logger.BaseLogger
}

func CreateAuthenticator(baseLogger logger.BaseLogger) BaseController {
	return &controller{
		log: baseLogger,
	}
}

func (c *controller) Route() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/register", c.registerNewUser)
	r.Post("/login", c.loginUser)

	return r
}

func (c *controller) registerNewUser(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func (c *controller) loginUser(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}
