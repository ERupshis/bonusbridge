package auth

import (
	"bytes"
	"net/http"

	"github.com/erupshis/bonusbridge/internal/auth/users/userdata"
	"github.com/erupshis/bonusbridge/internal/helpers"
)

//TODO: split in independent package.

func (c *Controller) loginHandler(w http.ResponseWriter, r *http.Request) {
	buf := bytes.Buffer{}
	if _, err := buf.ReadFrom(r.Body); err != nil {
		c.log.Info("[controller:loginHandler] failed to read request body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer helpers.ExecuteWithLogError(r.Body.Close, c.log)

	var user userdata.User
	if err := helpers.UnmarshalData(buf.Bytes(), &user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.log.Info("[controller:loginHandler] bad new user input userdata")
		return
	}

	userID, err := c.usersStrg.GetUserID(user.Login)
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
