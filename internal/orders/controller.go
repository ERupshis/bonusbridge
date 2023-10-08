package orders

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/erupshis/bonusbridge/internal/auth/users/userdata"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/storage"
	"github.com/erupshis/bonusbridge/internal/orders/validator"
	"github.com/go-chi/chi/v5"
)

const packageName = "orders"

type Controller struct {
	storage storage.Storage

	log logger.BaseLogger
}

func CreateController(storage storage.Storage, baseLogger logger.BaseLogger) Controller {
	return Controller{
		storage: storage,
		log:     baseLogger,
	}
}

func (c *Controller) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", c.addOrderHandler)
	r.Get("/", c.getOrdersHandler)
	return r
}

func (c *Controller) addOrderHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "text/plain" {
		c.log.Info("[%s:Controller:addOrderHandler] wrong body content type: %s", packageName, contentType)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var reqBody bytes.Buffer
	_, err := reqBody.ReadFrom(r.Body)
	if err != nil {
		c.log.Info("[%s:Controller:addOrderHandler] failed to read request's body: %v", packageName, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer helpers.ExecuteWithLogError(r.Body.Close, c.log)

	orderNumber := reqBody.String()
	if !validator.IsLuhnValid(orderNumber) {
		c.log.Info("[%s:Controller:addOrderHandler] order number didn't pass Luhn's algorithm check")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		c.log.Info("[%s:Controller:addOrderHandler] failed to extract userID: %v", packageName, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = c.storage.AddOrder(orderNumber, userID)
	if err != nil {
		if errors.As(err, &storage.ErrOrderWasAddedBefore) {
			c.log.Info("[%s:Controller:addOrderHandler] order '%s' has been already added by user '%d' before", packageName, orderNumber, userID)
			w.WriteHeader(http.StatusOK)
			return
		}

		if errors.As(err, &storage.ErrOrderWasAddedByAnotherUser) {
			c.log.Info("[%s:Controller:addOrderHandler] order '%s' has been already added by another user '%d' before", packageName, orderNumber, userID)
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	c.log.Info("[%s:Controller:addOrderHandler] order '%s' has been added in system. userID '%d'", packageName, orderNumber, userID)
	w.WriteHeader(http.StatusAccepted)
}

func (c *Controller) getOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		c.log.Info("[%s:Controller:addOrderHandler] failed to extract userID: %v", packageName, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	orders, err := c.storage.GetOrders(userID)
	if err != nil {
		c.log.Info("[%s:Controller:addOrderHandler] failed to get user's '%d' orders: %v", packageName, userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		c.log.Info("[%s:Controller:addOrderHandler] orders associated with user '%d' are not found: %v", packageName, userID, err)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	respBody, err := json.Marshal(orders)
	if err != nil {
		c.log.Info("[%s:Controller:addOrderHandler] failed to marshal user '%d' orders: %v", packageName, userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(respBody); err != nil {
		c.log.Info("[%s:Controller:addOrderHandler] failed to write orders data in response body: %v", packageName, err)
		w.WriteHeader(http.StatusInternalServerError)

	}
}

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
