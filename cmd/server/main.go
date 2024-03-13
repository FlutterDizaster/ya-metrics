package main

import (
	"net/http"

	"github.com/FlutterDizaster/ya-metrics/internal/handlers"
	"github.com/FlutterDizaster/ya-metrics/internal/storage"
)

func main() {
	storage := storage.NewMetricStorage()
	updateHandler := handlers.NewUpdateHandler(&storage)

	mux := http.NewServeMux()
	mux.Handle("/update/", updateHandler)

	http.ListenAndServe(":8080", mux)
}
