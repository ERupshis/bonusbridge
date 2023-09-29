package controllers

import (
	"github.com/go-chi/chi/v5"
)

type BaseController interface {
	Route() *chi.Mux
}
