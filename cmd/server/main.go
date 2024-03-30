package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/FlutterDizaster/ya-metrics/internal/handlers"
	"github.com/FlutterDizaster/ya-metrics/internal/memstorage"
)

func main() {
	endpoint := flag.String("a", "localhost:8080", "Server endpoint addres. Default localhost:8080")
	flag.Parse()

	envEndpoint, ok := os.LookupEnv("ADDRESS")
	if ok {
		endpoint = &envEndpoint
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

	http.ListenAndServe(*endpoint, handlers.NewRouter(rs))
}
