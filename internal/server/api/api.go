package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/go-chi/chi/v5"
	"golang.org/x/sync/errgroup"
)

// const (
// 	gauge   = "gauge"
// 	counter = "counter"
// )

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
	Addr        string
}

// API используется для обработки запросов к серверу.
// Для создания экземпляра необходимо испольщовать функцию New(*Settings) *API.
type API struct {
	storage MetricsStorage
	server  *http.Server
}

// Фабрика создания роутера.
// Необходима для правильной инициалищации экземпляра API.
func New(as *Settings) *API {
	slog.Debug("Creating API service")
	// создание экземпляра API
	api := &API{
		storage: as.Storage,
	}

	r := chi.NewRouter()

	// передача слайса Middleware функций в chi.Mux
	if as.Middlewares != nil {
		r.Use(as.Middlewares...)
	}

	// настройка роутинга
	r.Get("/", api.getAllHandler)
	r.Route("/update", func(rr chi.Router) {
		rr.Post("/", api.updateJSONHandler)
		rr.Post("/{kind}/{name}/{value}", api.updateHandler)
	})
	r.Route("/value", func(rr chi.Router) {
		rr.Post("/", api.getJSONMetricHandler)
		rr.Get("/{kind}/{name}", api.getMetricHandler)
	})

	// настройка ответов на не обрабатываемые сервером запросы
	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	api.server = &http.Server{
		Addr:    as.Addr,
		Handler: r,
	}
	slog.Debug("API service created")
	return api
}

func (api *API) Start(ctx context.Context) error {
	slog.Info("Starting API service")
	defer slog.Info("API server succesfully stopped")
	eg := errgroup.Group{}

	eg.Go(func() error {
		slog.Info("Listening...")
		err := api.server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	<-ctx.Done()
	eg.Go(func() error {
		slog.Info("Shutingdown API service")
		return api.server.Shutdown(context.TODO())
	})

	return eg.Wait()
}
