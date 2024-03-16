package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type RouterSettings struct {
	UpdateHandler    http.Handler
	GetAllHandler    http.Handler
	GetMetricHandler http.Handler
}

func NewRouter(rs RouterSettings) http.Handler {
	r := chi.NewRouter()

	r.Get("/", rs.GetAllHandler.ServeHTTP)                              // Main page handle
	r.Post("/update/{kind}/{name}/{value}", rs.UpdateHandler.ServeHTTP) // Update handle
	r.Get("/value/{kind}/{name}", rs.GetMetricHandler.ServeHTTP)        // Get value handle

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	return r
}
