package main

import (
	"flag"
	"os"

	"github.com/FlutterDizaster/ya-metrics/internal/server"
)

func main() {
	endpoint := flag.String("a", "localhost:8080", "Server endpoint addres. Default localhost:8080")
	flag.Parse()

	envEndpoint, ok := os.LookupEnv("ADDRESS")
	if ok {
		endpoint = &envEndpoint
	}

	server.Setup(*endpoint)
}
