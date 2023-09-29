package controllers

import (
	"github.com/go-chi/chi"
)

type BaseController interface {
	Route() *chi.Mux
}
