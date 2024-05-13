package agent

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/agent/sender"
	"github.com/FlutterDizaster/ya-metrics/internal/agent/telemetry"
	"github.com/FlutterDizaster/ya-metrics/internal/agent/worker"
	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/FlutterDizaster/ya-metrics/pkg/logger"
)

const (
	retryCount         = 10
	retryIntervalMS    = 200
	retryMaxWaitTime   = 2
	gracefullPeriodSec = 20
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

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	senderSettings := &sender.Settings{
		Addr:             endpoint,
		RetryCount:       retryCount,
		RetryInterval:    retryIntervalMS * time.Microsecond,
		RetryMaxWaitTime: retryMaxWaitTime * time.Second,
	}

	sender := sender.NewSender(senderSettings)

	collector := telemetry.NewMetricCollector(customMetricsList)

	workerSettings := &worker.Settings{
		Collector:      collector,
		Sender:         sender,
		ReportInterval: reportInterval,
		PollInterval:   pollInterval,
	}

	w, err := worker.NewWorker(workerSettings)
	if err != nil {
		slog.Error(
			"worker error",
			slog.String("error", err.Error()),
		)
	}

	var wg sync.WaitGroup

	// Создание контекста воркера
	workerCtx, workerCancleCtx := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		w.Start(workerCtx)
		wg.Done()
	}()

	// Ожидание завершения контекста сигналом системы
	<-ctx.Done()

	// Запуск Gracefull Keeper
	// Завершает выполнение программы через gracefullPeriodSec секунд, если программа не завершится сама
	forceCtx, forceStopCtx := context.WithTimeout(
		context.Background(),
		gracefullPeriodSec*time.Second,
	)
	defer forceStopCtx()
	go func() {
		<-forceCtx.Done()
		if forceCtx.Err() == context.DeadlineExceeded {
			slog.Error("shutdown timed out... forcing exit.")
			os.Exit(1)
		}
	}()

	workerCancleCtx()

	wg.Wait()
}
