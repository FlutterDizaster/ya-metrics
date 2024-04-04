package server

import (
	"errors"
	"log"
	"net/http"

	"github.com/FlutterDizaster/ya-metrics/internal/handlers"
	"github.com/FlutterDizaster/ya-metrics/internal/memstorage"
	"github.com/FlutterDizaster/ya-metrics/pkg/utils"
)

func Setup(url string) {
	// url validation
	err := utils.ValidateURL(url)
	if err != nil {
		log.Fatalf("url error: %s", err)
	}

	storage := memstorage.NewMetricStorage()

	updateHandler := handlers.NewUpdateHandler(&storage)
	getMetricHandler := handlers.NewGetMetricHandler(&storage)
	getAllHandler := handlers.NewGetAllHandler(&storage)

	rs := handlers.RouterSettings{
		UpdateHandler:    updateHandler,
		GetAllHandler:    getAllHandler,
		GetMetricHandler: getMetricHandler,
	}

	err = http.ListenAndServe(url, handlers.NewRouter(rs))
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %s", err)
	}
}
