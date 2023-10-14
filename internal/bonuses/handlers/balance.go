package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/erupshis/bonusbridge/internal/auth"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage"
	"github.com/erupshis/bonusbridge/internal/logger"
)

func Balance(storage storage.Storage, log logger.BaseLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.GetUserIDFromContext(r.Context())
		if err != nil {
			log.Info("[bonuses:handlers:Balance] failed to extract userID: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userBalance, err := storage.GetBalance(r.Context(), userID)
		if err != nil {
			log.Info("[bonuses:handlers:Balance] failed to get balance for userID '%d': %v", userID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		respBody, err := json.Marshal(userBalance)
		if err != nil {
			log.Info("[bonuses:handlers:Balance] failed to marshal userID's '%d' balance: %v", userID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Content-Length", fmt.Sprintf("%d", len(respBody)))
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write(respBody); err != nil {
			log.Info("[bonuses:handlers:Balance] failed to write orders data in response body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
