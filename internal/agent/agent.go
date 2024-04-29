package agent

import (
	"context"

	newmemory "github.com/FlutterDizaster/ya-metrics/internal/repository/new-memory"
	"github.com/FlutterDizaster/ya-metrics/internal/sender"
	"github.com/FlutterDizaster/ya-metrics/internal/telemetry"
	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/FlutterDizaster/ya-metrics/pkg/logger"
)

func Setup(endpoint string, reportInterval int, pollInterval int) {
	logger.Init()

	customMetricsList := []view.Metric{
		{ID: "Alloc", MType: "gauge", Source: view.MemStats},
		{ID: "BuckHashSys", MType: "gauge", Source: view.MemStats},
		{ID: "Frees", MType: "gauge", Source: view.MemStats},
		{ID: "GCCPUFraction", MType: "gauge", Source: view.MemStats},
		{ID: "GCSys", MType: "gauge", Source: view.MemStats},
		{ID: "HeapAlloc", MType: "gauge", Source: view.MemStats},
		{ID: "HeapIdle", MType: "gauge", Source: view.MemStats},
		{ID: "HeapInuse", MType: "gauge", Source: view.MemStats},
		{ID: "HeapObjects", MType: "gauge", Source: view.MemStats},
		{ID: "HeapReleased", MType: "gauge", Source: view.MemStats},
		{ID: "HeapSys", MType: "gauge", Source: view.MemStats},
		{ID: "LastGC", MType: "gauge", Source: view.MemStats},
		{ID: "Lookups", MType: "gauge", Source: view.MemStats},
		{ID: "MCacheInuse", MType: "gauge", Source: view.MemStats},
		{ID: "MCacheSys", MType: "gauge", Source: view.MemStats},
		{ID: "MSpanInuse", MType: "gauge", Source: view.MemStats},
		{ID: "MSpanSys", MType: "gauge", Source: view.MemStats},
		{ID: "Mallocs", MType: "gauge", Source: view.MemStats},
		{ID: "NextGC", MType: "gauge", Source: view.MemStats},
		{ID: "NumForcedGC", MType: "gauge", Source: view.MemStats},
		{ID: "NumGC", MType: "gauge", Source: view.MemStats},
		{ID: "OtherSys", MType: "gauge", Source: view.MemStats},
		{ID: "PauseTotalNs", MType: "gauge", Source: view.MemStats},
		{ID: "StackInuse", MType: "gauge", Source: view.MemStats},
		{ID: "StackSys", MType: "gauge", Source: view.MemStats},
		{ID: "Sys", MType: "gauge", Source: view.MemStats},
		{ID: "TotalAlloc", MType: "gauge", Source: view.MemStats},
	}

	storage := newmemory.NewMetricStorage()

	collector := telemetry.NewMetricCollector(storage, pollInterval, customMetricsList)

	sender := sender.NewSender(endpoint, reportInterval, storage)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go collector.Start(ctx)

	sender.Start(ctx)
}
