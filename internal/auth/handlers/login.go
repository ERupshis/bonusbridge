package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/erupshis/bonusbridge/internal/auth/jwtgenerator"
	"github.com/erupshis/bonusbridge/internal/auth/users/data"
	"github.com/erupshis/bonusbridge/internal/auth/users/managers"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
)

func Login(usersStorage managers.BaseUsersManager, jwt jwtgenerator.JwtGenerator, log logger.BaseLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		buf := bytes.Buffer{}
		if _, err := buf.ReadFrom(r.Body); err != nil {
			log.Info("[auth:handlers:Login] failed to read request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer helpers.ExecuteWithLogError(r.Body.Close, log)

		var user data.User
		if err := json.Unmarshal(buf.Bytes(), &user); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Info("[auth:handlers:Login] bad new user input data: %v", err)
			return
		}

		userDB, err := usersStorage.GetUser(r.Context(), user.Login)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Info("[auth:handlers:Login] failed to get userID from user's database: %v", err)
			return
		}

		if userDB == nil {
			w.WriteHeader(http.StatusUnauthorized)
			log.Info("[auth:handlers:Login] failed to get userID from user's database: %v", err)
			return
		}

		if user.Password != userDB.Password {
			w.WriteHeader(http.StatusUnauthorized)
			log.Info("[auth:handlers:Login] failed to authorize user")
			return
		}

		token, err := jwt.BuildJWTString(userDB.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Info("[auth:handlers:Login] new token generation failed: %w", err)
			return
		}

		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)

		log.Info("[auth:handlers:Login] user '%s' authenticated successfully", user.Login)
	}
}
