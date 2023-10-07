package auth

import (
	"bytes"
	"net/http"

	"github.com/erupshis/bonusbridge/internal/auth/users/data"
	"github.com/erupshis/bonusbridge/internal/helpers"
)

func (c *Controller) registerHandler(w http.ResponseWriter, r *http.Request) {
	buf := bytes.Buffer{}
	if _, err := buf.ReadFrom(r.Body); err != nil {
		c.log.Info("[controller:registerHandler] failed to read request body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer helpers.ExecuteWithLogError(r.Body.Close, c.log)

	var user data.User
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
