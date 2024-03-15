package main

import (
	"log"
	"net/http"

	"github.com/FlutterDizaster/ya-metrics/internal/handlers"
	"github.com/FlutterDizaster/ya-metrics/internal/storage"
)

func main() {
	storage := storage.NewMetricStorage()
	updateHandler := handlers.NewUpdateHandler(&storage)

	mux := http.NewServeMux()
	mux.Handle("/update/", updateHandler)

	log.Println("Listening")
	http.ListenAndServe(":8080", mux)
}
