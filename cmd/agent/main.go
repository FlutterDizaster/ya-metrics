package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/FlutterDizaster/ya-metrics/internal/memstorage"
	"github.com/FlutterDizaster/ya-metrics/internal/sender"
	"github.com/FlutterDizaster/ya-metrics/internal/telemetry"
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

	metricsList := []telemetry.Metric{
		{Name: "Alloc", Kind: telemetry.KindGauge},
		{Name: "BuckHashSys", Kind: telemetry.KindGauge},
		{Name: "Frees", Kind: telemetry.KindGauge},
		{Name: "GCCPUFraction", Kind: telemetry.KindGauge},
		{Name: "GCSys", Kind: telemetry.KindGauge},
		{Name: "HeapAlloc", Kind: telemetry.KindGauge},
		{Name: "HeapIdle", Kind: telemetry.KindGauge},
		{Name: "HeapInuse", Kind: telemetry.KindGauge},
		{Name: "HeapObjects", Kind: telemetry.KindGauge},
		{Name: "HeapReleased", Kind: telemetry.KindGauge},
		{Name: "HeapSys", Kind: telemetry.KindGauge},
		{Name: "LastGC", Kind: telemetry.KindGauge},
		{Name: "Lookups", Kind: telemetry.KindGauge},
		{Name: "MCacheInuse", Kind: telemetry.KindGauge},
		{Name: "MCacheSys", Kind: telemetry.KindGauge},
		{Name: "MSpanInuse", Kind: telemetry.KindGauge},
		{Name: "MSpanSys", Kind: telemetry.KindGauge},
		{Name: "Mallocs", Kind: telemetry.KindGauge},
		{Name: "NextGC", Kind: telemetry.KindGauge},
		{Name: "NumForcedGC", Kind: telemetry.KindGauge},
		{Name: "NumGC", Kind: telemetry.KindGauge},
		{Name: "OtherSys", Kind: telemetry.KindGauge},
		{Name: "PauseTotalNs", Kind: telemetry.KindGauge},
		{Name: "StackInuse", Kind: telemetry.KindGauge},
		{Name: "StackSys", Kind: telemetry.KindGauge},
		{Name: "Sys", Kind: telemetry.KindGauge},
		{Name: "TotalAlloc", Kind: telemetry.KindGauge},
		{Name: "PollCount", Kind: telemetry.KindCounter},
		{Name: "RandomValue", Kind: telemetry.KindGauge},
	}

	storage := memstorage.NewMetricStorage()

	collector := telemetry.NewMetricCollector(&storage, *pollInterval, metricsList)

	sender := sender.NewSender(*endpoint, *reportInterval, &storage)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go collector.Start(ctx)

	sender.Start(ctx)
}
