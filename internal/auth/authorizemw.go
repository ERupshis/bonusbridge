package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/erupshis/bonusbridge/internal/auth/users/userdata"
)

//TODO: split in independent package.

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

		ctxWithValue := context.WithValue(r.Context(), userdata.UserID, fmt.Sprintf("%d", userID))
		h.ServeHTTP(w, r.WithContext(ctxWithValue))
	})
}
