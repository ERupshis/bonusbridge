package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/erupshis/bonusbridge/internal/auth/users/userdata"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/storage"
	"github.com/erupshis/bonusbridge/internal/orders/validator"
)

const packageName = "orders"

func AddOrderHandler(strg storage.Storage, log logger.BaseLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType != "text/plain" {
			log.Info("[%s:AddOrderHandler] wrong body content type: %s", packageName, contentType)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var reqBody bytes.Buffer
		_, err := reqBody.ReadFrom(r.Body)
		if err != nil {
			log.Info("[%s:AddOrderHandler] failed to read request's body: %v", packageName, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer helpers.ExecuteWithLogError(r.Body.Close, log)

		orderNumber := reqBody.String()
		if !validator.IsLuhnValid(orderNumber) {
			log.Info("[%s:AddOrderHandler] order number didn't pass Luhn's algorithm check", packageName)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		userID, err := getUserIDFromContext(r.Context())
		if err != nil {
			log.Info("[%s:AddOrderHandler] failed to extract userID: %v", packageName, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = strg.AddOrder(r.Context(), orderNumber, userID)
		if err != nil {
			if errors.Is(err, storage.ErrOrderWasAddedBefore) {
				log.Info("[%s:AddOrderHandler] order '%s' has been already added by this user before", packageName, orderNumber, userID)
				w.WriteHeader(http.StatusOK)
				return
			}

			if errors.Is(err, storage.ErrOrderWasAddedByAnotherUser) {
				log.Info("[%s:AddOrderHandler] order '%s' has been already added by another user before", packageName, orderNumber, userID)
				w.WriteHeader(http.StatusConflict)
				return
			}

			log.Info("[%s:AddOrderHandler] unknown error: %v", packageName, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Info("[%s:AddOrderHandler] order '%s' has been added in system. userID '%d'", packageName, orderNumber, userID)
		w.WriteHeader(http.StatusAccepted)
	}
}

// TODO: move in auth's helpers.
func getUserIDFromContext(ctx context.Context) (int64, error) {
	userIDraw := ctx.Value(userdata.UserID)
	if userIDraw == nil {
		return -1, fmt.Errorf("missing userID in request's context")
	}

	userID, err := strconv.ParseInt(userIDraw.(string), 10, 64)
	if err != nil {
		return -1, fmt.Errorf("parse userID from request's context: %w", err)
	}

	return userID, nil
}
