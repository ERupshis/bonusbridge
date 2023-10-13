package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erupshis/bonusbridge/internal/auth"
	"github.com/erupshis/bonusbridge/internal/bonuses/data"
	"github.com/erupshis/bonusbridge/internal/bonuses/storage"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/validator"
)

func Withdraw(strg storage.Storage, log logger.BaseLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.GetUserIDFromContext(r.Context())
		if err != nil {
			log.Info("[bonuses:handlers:Withdraw] failed to extract userID: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		buf := bytes.Buffer{}
		if _, err := buf.ReadFrom(r.Body); err != nil {
			log.Info("[bonuses:handlers:Withdraw] failed to read request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer helpers.ExecuteWithLogError(r.Body.Close, log)

		var withdrawal data.Withdrawal
		if err = json.Unmarshal(buf.Bytes(), &withdrawal); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Info("[bonuses:handlers:Withdraw] failed to unmarshal request body: %v", err)
			return
		}

		if !validator.IsLuhnValid(withdrawal.Order) {
			log.Info("[bonuses:handlers:Withdraw] order number didn't pass Luhn's algorithm check")
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		withdrawal.UserID = userID
		if err = strg.WithdrawBonuses(r.Context(), &withdrawal); err != nil {
			if errors.Is(err, storage.ErrNotEnoughBonuses) {
				w.WriteHeader(http.StatusPaymentRequired)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			log.Info("[bonuses:handlers:Withdraw] failed to withdraw bonuses: %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
