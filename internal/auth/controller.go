package auth

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/erupshis/bonusbridge/internal/auth/jwtgenerator"
	"github.com/erupshis/bonusbridge/internal/auth/users"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/go-chi/chi/v5"
)

const packageName = "auth"

type Controller struct {
	usersStrg users.Storage
	jwt       jwtgenerator.JwtGenerator

	log logger.BaseLogger
}

func CreateAuthenticator(usersStorage users.Storage, jwt jwtgenerator.JwtGenerator, baseLogger logger.BaseLogger) *Controller {
	return &Controller{
		usersStrg: usersStorage,
		jwt:       jwt,
		log:       baseLogger,
	}
}

func (c *Controller) Route() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/register", c.registerHandler)
	r.Post("/login", c.loginHandler)

	return r
}

func (c *Controller) registerHandler(w http.ResponseWriter, r *http.Request) {
	buf := bytes.Buffer{}
	if _, err := buf.ReadFrom(r.Body); err != nil {
		c.log.Info("[controller:registerHandler] failed to read request body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer helpers.ExecuteWithLogError(r.Body.Close, c.log)

	var user users.User
	if err := helpers.UnmarshalData(buf.Bytes(), &user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.log.Info("[controller:registerHandler] bad new user input data")
		return
	}

	userID, err := c.usersStrg.GetUserId(user.Login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.log.Info("[controller:registerHandler] failed to check user in database")
		return
	}

	if userID != -1 {
		w.WriteHeader(http.StatusConflict)
		c.log.Info("[controller:registerHandler] login already exists")
		return
	}

	userID, err = c.usersStrg.AddUser(user.Login, user.Password)
	if err != nil || userID == -1 {
		w.WriteHeader(http.StatusInternalServerError)
		c.log.Info("[controller:registerHandler] failed to add new user '%s'", user.Login)
		return
	}

	token, err := c.jwt.BuildJWTString(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.log.Info("[controller:loginHandler] new token generation failed: %w", err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusCreated)

	c.log.Info("[controller:registerHandler] user '%s' registered successfully", user.Login)
}

func (c *Controller) loginHandler(w http.ResponseWriter, r *http.Request) {
	buf := bytes.Buffer{}
	if _, err := buf.ReadFrom(r.Body); err != nil {
		c.log.Info("[controller:loginHandler] failed to read request body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer helpers.ExecuteWithLogError(r.Body.Close, c.log)

	var user users.User
	if err := helpers.UnmarshalData(buf.Bytes(), &user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.log.Info("[controller:loginHandler] bad new user input data")
		return
	}

	userID, err := c.usersStrg.GetUserId(user.Login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.log.Info("[controller:loginHandler] failed to get userID from user's database: %w", err)
		return
	}

	if userID == -1 {
		w.WriteHeader(http.StatusUnauthorized)
		c.log.Info("[controller:loginHandler] failed to get userID from user's database: %w", err)
		return
	}

	authorized, err := c.usersStrg.ValidateUser(user.Login, user.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.log.Info("[controller:loginHandler] failed to check user's login/password in database")
		return
	}

	if !authorized {
		w.WriteHeader(http.StatusUnauthorized)
		c.log.Info("[controller:loginHandler] failed to authorize user")
		return
	}

	token, err := c.jwt.BuildJWTString(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.log.Info("[controller:loginHandler] new token generation failed: %w", err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)

	c.log.Info("[controller:registerHandler] user '%s' authenticated successfully", user.Login)
}

func (c *Controller) AuthorizeUser(h http.Handler, userRoleRequirement int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			c.log.Info("[%s:controller:AuthorizeUser] invalid request without authentication token", packageName)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token := strings.Split(authHeader, " ")
		if len(token) != 2 || token[0] != "Bearer" {
			c.log.Info("[%s:controller:AuthorizeUser] invalid invalid token", packageName)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID := c.jwt.GetUserId(token[1])
		userRole, err := c.usersStrg.GetUserRole(userID)
		if err != nil {
			c.log.Info("[%s:controller:AuthorizeUser] failed to search user in system: %v", packageName, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if userRole == -1 {
			c.log.Info("[%s:controller:AuthorizeUser] user is not registered in system", packageName)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if userRole < userRoleRequirement {
			c.log.Info("[%s:controller:AuthorizeUser] user doesn't have permission to resource: %s", packageName, r.URL.Path)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		h.ServeHTTP(w, r)
	})
}
