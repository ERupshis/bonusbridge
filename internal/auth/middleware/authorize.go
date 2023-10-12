package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/erupshis/bonusbridge/internal/auth/jwtgenerator"
	"github.com/erupshis/bonusbridge/internal/auth/users/data"
	"github.com/erupshis/bonusbridge/internal/auth/users/managers"
	"github.com/erupshis/bonusbridge/internal/logger"
)

type ContextString string

func AuthorizeUser(h http.Handler, userRoleRequirement int, usersStorage managers.BaseUsersManager, jwt jwtgenerator.JwtGenerator, log logger.BaseLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Info("[auth:middleware:Authorize] invalid request without authentication token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token := strings.Split(authHeader, " ")
		if len(token) != 2 || token[0] != "Bearer" {
			log.Info("[auth:middleware:Authorize] invalid invalid token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID := jwt.GetUserID(token[1])
		userRole, err := usersStorage.GetUserRole(userID)
		if err != nil {
			log.Info("[auth:middleware:Authorize] failed to search user in system: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if userRole == -1 {
			log.Info("[auth:middleware:Authorize] user is not registered in system")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if userRole < userRoleRequirement {
			log.Info("[auth:middleware:Authorize] user doesn't have permission to resource: %s", r.URL.Path)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		ctxWithValue := context.WithValue(r.Context(), ContextString(data.UserID), fmt.Sprintf("%d", userID))
		h.ServeHTTP(w, r.WithContext(ctxWithValue))
	})
}
