package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/erupshis/bonusbridge/internal/auth"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
)

func Balance(storage storage.Storage, log logger.BaseLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		buf := bytes.Buffer{}
		if _, err := buf.ReadFrom(r.Body); err != nil {
			log.Info("[bonuses:handlers:Balance] failed to read request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer helpers.ExecuteWithLogError(r.Body.Close, log)

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

		//TODO: check necessity.
		if userBalance == nil {
			log.Info("[bonuses:handlers:Balance] failed to get balance: null")
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
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write(respBody); err != nil {
			log.Info("[bonuses:handlers:Balance] failed to write orders data in response body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
