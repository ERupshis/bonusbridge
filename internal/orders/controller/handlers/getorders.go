package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/storage"
)

func GetOrdersHandler(strg storage.Storage, log logger.BaseLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserIDFromContext(r.Context())
		if err != nil {
			log.Info("[%s:addOrderHandler] failed to extract userID: %v", packageName, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		orders, err := strg.GetOrders(r.Context(), userID)
		if err != nil {
			log.Info("[%s:addOrderHandler] failed to get user's '%d' orders: %v", packageName, userID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(orders) == 0 {
			log.Info("[%s:addOrderHandler] orders associated with user '%d' are not found: %v", packageName, userID, err)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		respBody, err := json.Marshal(orders)
		if err != nil {
			log.Info("[%s:addOrderHandler] failed to marshal user '%d' orders: %v", packageName, userID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write(respBody); err != nil {
			log.Info("[%s:addOrderHandler] failed to write orders data in response body: %v", packageName, err)
			w.WriteHeader(http.StatusInternalServerError)

		}
	}
}
