package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/FlutterDizaster/ya-metrics/internal/agent"
)

func main() {
	endpoint := flag.String("a", "localhost:8080", "HTTP-server addres. Default \"localhost:8080\"")
	reportInterval := flag.Int("r", 10, "Report interval in seconds. Default 10 sec")
	pollInterval := flag.Int("p", 2, "Metrics poll interval. Default 2 sec")
	flag.Parse()

	func() {
		envEndpoint, ok := os.LookupEnv("ADDRESS")
		if ok {
			endpoint = &envEndpoint
		}

		envReportInterval, ok := os.LookupEnv("REPORT_INTERVAL")
		if ok {
			rInerval, err := strconv.Atoi(envReportInterval)
			if err != nil {
				log.Fatalln(err)
			}
			reportInterval = &rInerval
		}

		envPollInterval, ok := os.LookupEnv("POLL_INTERVAL")
		if ok {
			pInterval, err := strconv.Atoi(envPollInterval)
			if err != nil {
				log.Fatalln(err)
			}
			pollInterval = &pInterval
		}
	}()

	agent.Setup(*endpoint, *reportInterval, *pollInterval)
}
