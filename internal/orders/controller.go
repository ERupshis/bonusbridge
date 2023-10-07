package orders

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/erupshis/bonusbridge/internal/auth/users/data"
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
		c.log.Info("[%s:Controller:addOrderHandler] wrong body content type: %s", contentType)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var reqBody []byte
	_, err := r.Body.Read(reqBody)
	if err != nil {
		c.log.Info("[%s:Controller:addOrderHandler] failed to read request's body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer helpers.ExecuteWithLogError(r.Body.Close, c.log)

	orderNumber := string(reqBody)
	if !validator.IsLuhnValid(orderNumber) {
		c.log.Info("[%s:Controller:addOrderHandler] order number didn't pass Luhn's algorithm check")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	userIDstring := r.Header.Get(data.UserID)
	userID, err := strconv.ParseInt(userIDstring, 10, 64)
	if err != nil {
		c.log.Info("[%s:Controller:addOrderHandler] failed to parse userID from request header: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = c.storage.AddOrder(orderNumber, userID)
	if err != nil {
		if errors.As(err, &storage.ErrOrderWasAddedBefore) {
			c.log.Info("[%s:Controller:addOrderHandler] order '%s' has been already added by user '%d' before", orderNumber, userID)
			w.WriteHeader(http.StatusOK)
			return
		}

		if errors.As(err, &storage.ErrOrderWasAddedByAnotherUser) {
			c.log.Info("[%s:Controller:addOrderHandler] order '%s' has been already added by another user '%d' before", orderNumber, userID)
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	c.log.Info("[%s:Controller:addOrderHandler] order '%s' has been added in system. userID '%d'", orderNumber, userID)
	w.WriteHeader(http.StatusAccepted)
}

func (c *Controller) getOrdersHandler(w http.ResponseWriter, r *http.Request) {

}
