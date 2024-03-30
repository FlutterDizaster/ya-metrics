package main

import (
	"net/http"

	"github.com/FlutterDizaster/ya-metrics/internal/handlers"
	"github.com/FlutterDizaster/ya-metrics/internal/storage"
)

func main() {
	storage := storage.NewMetricStorage()

	updateHandler := handlers.NewUpdateHandler(&storage)
	getMetricHandler := handlers.NewGetMetricHandler(&storage)
	getAllHandler := handlers.NewGetAllHandler(&storage)

	rs := handlers.RouterSettings{
		UpdateHandler:    updateHandler,
		GetAllHandler:    getAllHandler,
		GetMetricHandler: getMetricHandler,
	}

	http.ListenAndServe(":8080", handlers.NewRouter(rs))
}
