package auth

import (
	"net/http"

	"github.com/erupshis/bonusbridge/internal/auth/handlers"
	"github.com/erupshis/bonusbridge/internal/auth/jwtgenerator"
	"github.com/erupshis/bonusbridge/internal/auth/middleware"
	"github.com/erupshis/bonusbridge/internal/auth/users/managers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/go-chi/chi/v5"
)

type Controller struct {
	usersStrg managers.BaseUsersManager
	jwt       jwtgenerator.JwtGenerator

	log logger.BaseLogger
}

func CreateController(usersStorage managers.BaseUsersManager, jwt jwtgenerator.JwtGenerator, baseLogger logger.BaseLogger) *Controller {
	return &Controller{
		usersStrg: usersStorage,
		jwt:       jwt,
		log:       baseLogger,
	}
}

func (c *Controller) RouteRegister() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", handlers.Register(c.usersStrg, c.jwt, c.log))
	return r
}

func (c *Controller) RouteLoginer() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", handlers.Login(c.usersStrg, c.jwt, c.log))
	return r
}

func (c *Controller) AuthorizeUser(h http.Handler, userRoleRequirement int) http.Handler {
	return middleware.AuthorizeUser(h, userRoleRequirement, c.usersStrg, c.jwt, c.log)
}
