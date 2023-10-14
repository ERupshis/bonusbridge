package handlers

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/erupshis/bonusbridge/internal/auth"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
	"github.com/erupshis/bonusbridge/internal/orders/storage"
	"github.com/erupshis/bonusbridge/internal/orders/validator"
)

func AddOrderHandler(strg storage.Storage, log logger.BaseLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType != "text/plain" {
			log.Info("[orders:handlers:AddOrderHandler] wrong body content type: %s", contentType)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var reqBody bytes.Buffer
		_, err := reqBody.ReadFrom(r.Body)
		if err != nil {
			log.Info("[orders:handlers:AddOrderHandler] failed to read request's body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer helpers.ExecuteWithLogError(r.Body.Close, log)

		orderNumber := reqBody.String()
		if !validator.IsLuhnValid(orderNumber) {
			log.Info("[orders:handlers:AddOrderHandler] order number didn't pass Luhn's algorithm check")
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		userID, err := auth.GetUserIDFromContext(r.Context())
		if err != nil {
			log.Info("[orders:handlers:AddOrderHandler] failed to extract userID: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = strg.AddOrder(r.Context(), orderNumber, userID)
		if err != nil {
			if errors.Is(err, data.ErrOrderWasAddedBefore) {
				log.Info("[orders:handlers:AddOrderHandler] order '%s' has been already added by this user before", orderNumber)
				w.WriteHeader(http.StatusOK)
				return
			}

			if errors.Is(err, data.ErrOrderWasAddedByAnotherUser) {
				log.Info("[orders:handlers:AddOrderHandler] order '%s' has been already added by another user before", orderNumber)
				w.WriteHeader(http.StatusConflict)
				return
			}

			log.Info("[orders:handlers:AddOrderHandler] unknown error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Info("[orders:handlers:AddOrderHandler] order '%s' has been added in system. userID '%d'", orderNumber, userID)
		w.WriteHeader(http.StatusAccepted)
	}
}
