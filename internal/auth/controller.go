package auth

import (
	"github.com/erupshis/bonusbridge/internal/auth/jwtgenerator"
	"github.com/erupshis/bonusbridge/internal/auth/users/managers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/go-chi/chi/v5"
)

const packageName = "auth"

type Controller struct {
	usersStrg managers.BaseUsersManager
	jwt       jwtgenerator.JwtGenerator

	log logger.BaseLogger
}

func CreateAuthenticator(usersStorage managers.BaseUsersManager, jwt jwtgenerator.JwtGenerator, baseLogger logger.BaseLogger) *Controller {
	return &Controller{
		usersStrg: usersStorage,
		jwt:       jwt,
		log:       baseLogger,
	}
}

func (c *Controller) RouteRegister() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", c.registerHandler)
	return r
}

func (c *Controller) RouteLoginer() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", c.loginHandler)
	return r
}
