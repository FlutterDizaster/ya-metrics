package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"

	"github.com/FlutterDizaster/ya-metrics/internal/server/api/middleware"
	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/go-chi/chi/v5"
)

// Интерфейс взаимодействия с репозиторием метрик.
type MetricsStorage interface {
	AddMetrics(...view.Metric) ([]view.Metric, error)
	GetMetric(kind string, name string) (view.Metric, error)
	ReadAllMetrics() ([]view.Metric, error)
	Ping() error
}

// Структура Settings хранит параметры необходимые для создания экземпляра Router.
// Storage принимает репозиторий реалищующий интерфейс MetricsStorage.
// Middlewares принимает слайс Middleware функций соответствующих сигнатуре func(http.Handler) http.Handler.
// Middlewares может иметь значение nil.
type Settings struct {
	Storage     MetricsStorage
	Middlewares []middleware.Middleware
	Addr        string
}

// API используется для обработки запросов к серверу.
// Для создания экземпляра необходимо испольщовать функцию New(*Settings) *API.
type API struct {
	storage MetricsStorage
	server  *http.Server
}

// Фабрика создания API.
// Необходима для правильной инициалищации экземпляра API.
func New(as *Settings) *API {
	slog.Debug("Creating API service")
	// создание экземпляра API
	api := &API{
		storage: as.Storage,
	}

	r := chi.NewRouter()

	// передача Middleware функций в chi.Mux
	for i := range as.Middlewares {
		r.Use(as.Middlewares[i].Handle)
	}
	// r.Use(as.Middlewares...)

	// настройка роутинга
	r.Get("/", api.getAllHandler)
	r.Get("/ping", api.pingHandler)
	r.Post("/updates/", api.updateBatchHandler)
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

// Функция запуска сервсиса.
// TODO: Избавиться от errgroup
func (api *API) Start(ctx context.Context) error {
	slog.Info("Starting API service")
	defer slog.Info("API server succesfully stopped")
	wg := sync.WaitGroup{}

	errCh := make(chan error)
	defer close(errCh)

	wg.Add(1)
	go func() {
		api.start(errCh)
		wg.Done()
	}()

	var err error

	select {
	case <-ctx.Done():
		api.stop(errCh)
	case err = <-errCh:
		return err
	}

	return <-errCh
}

func (api *API) start(errCh chan error) {
	slog.Info("Listening...")
	err := api.server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		errCh <- err
	} else {
		errCh <- nil
	}
}

func (api *API) stop(errCh chan error) {
	slog.Info("Shutingdown API service")
	errCh <- api.server.Shutdown(context.TODO())
}
