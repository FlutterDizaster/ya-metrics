package server

import (
	"errors"
	"log"
	"log/slog"
	"net/http"

	"github.com/FlutterDizaster/ya-metrics/internal/logger"
	"github.com/FlutterDizaster/ya-metrics/internal/memstorage"
	"github.com/FlutterDizaster/ya-metrics/internal/router"
	"github.com/FlutterDizaster/ya-metrics/internal/router/middleware"
	"github.com/FlutterDizaster/ya-metrics/pkg/utils"
)

func Setup(url string) {
	// initialize logger
	logger.Init()

	// validate url
	if err := utils.ValidateURL(url); err != nil {
		log.Fatalf("url error: %s", err)
	}

	// create new metric storage
	storage := memstorage.NewMetricStorage()

	// configure router settings
	routerSettings := &router.Settings{
		Storage:     &storage,
		Middlewares: []func(http.Handler) http.Handler{middleware.Logger},
	}

	// start http server
	err := http.ListenAndServe(url, router.NewRouter(routerSettings))
	if !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server error: %s", err)
	}
}
