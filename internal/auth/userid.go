package auth

import (
	"context"
	"fmt"
	"strconv"

	"github.com/erupshis/bonusbridge/internal/auth/middleware"
	"github.com/erupshis/bonusbridge/internal/auth/users/data"
)

func GetUserIDFromContext(ctx context.Context) (int64, error) {
	userIDraw := ctx.Value(middleware.ContextString(data.UserID))
	if userIDraw == nil {
		return -1, fmt.Errorf("missing userID in request's context")
	}

	userID, err := strconv.ParseInt(userIDraw.(string), 10, 64)
	if err != nil {
		return -1, fmt.Errorf("parse userID from request's context: %w", err)
	}

	return userID, nil
}
