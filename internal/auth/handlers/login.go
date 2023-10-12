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
		if err := helpers.UnmarshalData(buf.Bytes(), &user); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Info("[auth:handlers:Login] bad new user input data")
			return
		}

		userID, err := usersStorage.GetUserID(user.Login)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Info("[auth:handlers:Login] failed to get userID from user's database: %w", err)
			return
		}

		if userID == -1 {
			w.WriteHeader(http.StatusUnauthorized)
			log.Info("[auth:handlers:Login] failed to get userID from user's database: %w", err)
			return
		}

		authorized, err := usersStorage.ValidateUser(user.Login, user.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Info("[auth:handlers:Login] failed to check user's login/password in database")
			return
		}

		if !authorized {
			w.WriteHeader(http.StatusUnauthorized)
			log.Info("[auth:handlers:Login] failed to authorize user")
			return
		}

		token, err := jwt.BuildJWTString(userID)
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
