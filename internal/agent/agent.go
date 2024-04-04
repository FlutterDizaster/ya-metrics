package agent

import (
	"context"

	"github.com/FlutterDizaster/ya-metrics/internal/memstorage"
	"github.com/FlutterDizaster/ya-metrics/internal/sender"
	"github.com/FlutterDizaster/ya-metrics/internal/telemetry"
)

func Setup(endpoint string, reportInterval int, pollInterval int) {
	customMetricsList := []telemetry.Metric{
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
	}

	storage := memstorage.NewMetricStorage()

	collector := telemetry.NewMetricCollector(&storage, pollInterval, customMetricsList)

	sender := sender.NewSender(endpoint, reportInterval, &storage)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go collector.Start(ctx)

	sender.Start(ctx)
}
