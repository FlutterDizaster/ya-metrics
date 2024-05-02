package router

import (
	"net/http"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/go-chi/chi/v5"
)

const (
	gauge   = "gauge"
	counter = "counter"
)

// Интерфейс взаимодействия с репозиторием метрик.
type MetricsStorage interface {
	AddMetric(view.Metric) (view.Metric, error)
	GetMetric(kind string, name string) (view.Metric, error)
	ReadAllMetrics() []view.Metric
}

// Структура Settings хранит параметры необходимые для создания экземпляра Router.
// Storage принимает репозиторий реалищующий интерфейс MetricsStorage.
// Middlewares принимает слайс Middleware функций соответствующих сигнатуре func(http.Handler) http.Handler.
// Middlewares может иметь значение nil.
type Settings struct {
	Storage     MetricsStorage
	Middlewares []func(http.Handler) http.Handler
}

// Router используется для обработки запросов к серверу.
// Для создания экземпляра необходимо испольщовать функцию NewRouter(*Settings) *Router.
type Router struct {
	*chi.Mux
	storage MetricsStorage
}

// Фабрика создания роутера.
// Необзодима для правильной инициалищации экземпляра Router.
func NewRouter(rs *Settings) *Router {
	// создание экземпляра Router
	r := &Router{
		Mux:     chi.NewRouter(),
		storage: rs.Storage,
	}

	// передача слайса Middleware функций в chi.Mux
	if rs.Middlewares != nil {
		r.Use(rs.Middlewares...)
	}

	// настройка роутинга
	r.Get("/", r.getAllHandler)
	r.Route("/update", func(rr chi.Router) {
		rr.Post("/", r.updateJSONHandler)
		rr.Post("/{kind}/{name}/{value}", r.updateHandler)
	})
	r.Route("/value", func(rr chi.Router) {
		rr.Post("/", r.getJSONMetricHandler)
		rr.Get("/{kind}/{name}", r.getMetricHandler)
	})

	// настройка ответов на не обрабатываемые сервером запросы
	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	return r
}
