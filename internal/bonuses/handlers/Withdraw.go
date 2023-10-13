package handlers

import (
	"bytes"
	"net/http"

	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
)

func Withdraw(log logger.BaseLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		buf := bytes.Buffer{}
		if _, err := buf.ReadFrom(r.Body); err != nil {
			log.Info("[bonuses:handlers:Withdraw] failed to read request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer helpers.ExecuteWithLogError(r.Body.Close, log)

		//TODO: need to add implementation.

		w.WriteHeader(http.StatusOK)
	}
}
