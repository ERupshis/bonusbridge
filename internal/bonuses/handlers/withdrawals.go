package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/erupshis/bonusbridge/internal/auth"
	"github.com/erupshis/bonusbridge/internal/bonuses/data"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage"
	"github.com/erupshis/bonusbridge/internal/logger"
)

func Withdrawals(strg storage.BaseBonusesStorage, log logger.BaseLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.GetUserIDFromContext(r.Context())
		if err != nil {
			log.Info("[bonuses:handlers:Withdrawals] failed to extract userID: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		withdrawals, err := strg.GetWithdrawals(r.Context(), userID)
		if err != nil {
			if errors.Is(err, data.ErrWithdrawalsMissing) {
				w.WriteHeader(http.StatusNoContent)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			log.Info("[bonuses:handlers:Withdrawals] failed to get withdrawals: %v", err)
			return
		}

		respBody, err := json.Marshal(withdrawals)
		if err != nil {
			log.Info("[bonuses:handlers:Withdrawals] failed convert withdrawals to JSON: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length-Type", fmt.Sprintf("%d", len(respBody)))
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(respBody); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
