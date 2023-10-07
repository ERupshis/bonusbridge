package orders

import (
	"net/http"

	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/go-chi/chi/v5"
)

type Controller struct {
	log logger.BaseLogger
}

func CreateController(baseLogger logger.BaseLogger) Controller {
	return Controller{
		log: baseLogger,
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
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//200 — номер заказа уже был загружен этим пользователем;
	//202 — новый номер заказа принят в обработку;
	//409 — номер заказа уже был загружен другим пользователем;
	//422 — неверный формат номера заказа;
	//500 — внутренняя ошибка сервера.
}

func (c *Controller) getOrdersHandler(w http.ResponseWriter, r *http.Request) {

}
