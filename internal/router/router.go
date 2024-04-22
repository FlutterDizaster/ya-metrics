package router

import (
	"net/http"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/go-chi/chi/v5"
)

type MetricsStorage interface {
	AddMetricValue(kind string, name string, value string) error
	GetMetricValue(kind string, name string) (string, error)
	ReadAllMetrics() []view.Metric
}

type Settings struct {
	Storage     MetricsStorage
	Middlewares []func(http.Handler) http.Handler
}

type Router struct {
	*chi.Mux
	storage MetricsStorage
}

func NewRouter(rs *Settings) *Router {
	r := &Router{
		Mux:     chi.NewRouter(),
		storage: rs.Storage,
	}

	if rs.Middlewares != nil {
		r.Use(rs.Middlewares...)
	}

	r.Get("/", r.getAllHandler)
	r.Post("/update/{kind}/{name}/{value}", r.updateHandler)
	r.Get("/value/{kind}/{name}", r.getMetricHandler)

	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	return r
}
