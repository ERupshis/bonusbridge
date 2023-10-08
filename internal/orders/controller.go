package orders

import (
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

	var reqBody []byte
	_, err := r.Body.Read(reqBody)
	if err != nil {
		c.log.Info("[%s:Controller:addOrderHandler] failed to read request's body: %v", packageName, err)
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

	//200 — успешная обработка запроса.
	//204 — нет данных для ответа.
	//401 — пользователь не авторизован.
	//500 — внутренняя ошибка сервера.
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
