package handlers

import (
	"bytes"
	"net/http"

	"github.com/erupshis/bonusbridge/internal/auth/jwtgenerator"
	"github.com/erupshis/bonusbridge/internal/auth/users/data"
	"github.com/erupshis/bonusbridge/internal/auth/users/managers"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
)

func Register(usersStorage managers.BaseUsersManager, jwt jwtgenerator.JwtGenerator, log logger.BaseLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		buf := bytes.Buffer{}
		if _, err := buf.ReadFrom(r.Body); err != nil {
			log.Info("[auth:handlers:Register] failed to read request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer helpers.ExecuteWithLogError(r.Body.Close, log)

		var user data.User
		if err := helpers.UnmarshalData(buf.Bytes(), &user); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Info("[auth:handlers:Register] bad new user input data")
			return
		}

		userID, err := usersStorage.GetUserID(user.Login)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Info("[auth:handlers:Register] failed to check user in database")
			return
		}

		if userID != -1 {
			w.WriteHeader(http.StatusConflict)
			log.Info("[auth:handlers:Register] login already exists")
			return
		}

		userID, err = usersStorage.AddUser(user.Login, user.Password)
		if err != nil || userID == -1 {
			w.WriteHeader(http.StatusInternalServerError)
			log.Info("[auth:handlers:Register] failed to add new user '%s'", user.Login)
			return
		}

		token, err := jwt.BuildJWTString(userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Info("[auth:handlers:Register] new token generation failed: %w", err)
			return
		}

		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)

		log.Info("[auth:handlers:Register] user '%s' registered successfully", user.Login)
	}
}
