package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

//TODO: Реализовать Telemetry

type Buffer interface {
	Put([]view.Metric) error
	Close()
}

type Settings struct {
	PollInterval time.Duration
	Buf          Buffer
}

type Telemetry struct {
	pollInterval time.Duration
	buf          Buffer
	metrics      []view.Metric
	rnd          rand.Rand
}

func New(settings Settings) *Telemetry {
	randSource := rand.NewSource(time.Now().UnixNano())
	return &Telemetry{
		pollInterval: settings.PollInterval,
		buf:          settings.Buf,
		rnd:          *rand.New(randSource),
	}
}

func (t *Telemetry) Start(ctx context.Context) error {
	slog.Debug("Telemetry", slog.String("status", "start"))
	ticker := time.NewTicker(t.pollInterval)

	// Первая сборка метрик
	t.collectAndSave()

	// Старт основного цикла
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			t.collectAndSave()
			t.buf.Close()
			slog.Debug("Telemetry", slog.String("status", "stop"))
			return nil
		case <-ticker.C:
			t.collectAndSave()
		}
	}
}

func (t *Telemetry) collectAndSave() {
	slog.Debug("Telemetry", slog.String("status", "collecting..."))
	// Очистка внутреннего буфера
	t.metrics = make([]view.Metric, 0)

	// Получение PollCount
	t.saveMetric(view.KindCounter, "PollCount", "1")

	// Получение RandomValue
	rvalue := strconv.FormatFloat(t.rnd.Float64(), 'f', -1, 64)
	t.saveMetric(view.KindGauge, "RandomValue", rvalue)

	// Получение метрик MemStats
	t.collectMemStats()

	// TODO: Получение метрик из других источников

	// Отправка метрик в буффер приложения
	err := t.buf.Put(t.metrics)
	if err != nil {
		slog.Error("Telemetry", "error", err)
		return
	}

	slog.Debug("Telemetry", slog.String("status", "collected"))
}

func (t *Telemetry) saveMetric(kind string, name string, value string) {
	// создание метрики
	metric, err := view.NewMetric(kind, name, value)
	if err != nil {
		slog.Error("%s metric not created: %s", name, err)
		return
	}
	// добавление метрики в буффер
	t.metrics = append(t.metrics, *metric)
}

func (t *Telemetry) collectMemStats() {
	memStatsMetrics := []view.Metric{
		{ID: "Alloc", MType: "gauge"},
		{ID: "BuckHashSys", MType: "gauge"},
		{ID: "Frees", MType: "gauge"},
		{ID: "GCCPUFraction", MType: "gauge"},
		{ID: "GCSys", MType: "gauge"},
		{ID: "HeapAlloc", MType: "gauge"},
		{ID: "HeapIdle", MType: "gauge"},
		{ID: "HeapInuse", MType: "gauge"},
		{ID: "HeapObjects", MType: "gauge"},
		{ID: "HeapReleased", MType: "gauge"},
		{ID: "HeapSys", MType: "gauge"},
		{ID: "LastGC", MType: "gauge"},
		{ID: "Lookups", MType: "gauge"},
		{ID: "MCacheInuse", MType: "gauge"},
		{ID: "MCacheSys", MType: "gauge"},
		{ID: "MSpanInuse", MType: "gauge"},
		{ID: "MSpanSys", MType: "gauge"},
		{ID: "Mallocs", MType: "gauge"},
		{ID: "NextGC", MType: "gauge"},
		{ID: "NumForcedGC", MType: "gauge"},
		{ID: "NumGC", MType: "gauge"},
		{ID: "OtherSys", MType: "gauge"},
		{ID: "PauseTotalNs", MType: "gauge"},
		{ID: "StackInuse", MType: "gauge"},
		{ID: "StackSys", MType: "gauge"},
		{ID: "Sys", MType: "gauge"},
		{ID: "TotalAlloc", MType: "gauge"},
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	for i := range memStatsMetrics {
		field := reflect.ValueOf(memStats).FieldByName((memStatsMetrics[i].ID))
		if field.IsValid() {
			t.saveMetric(
				memStatsMetrics[i].MType,
				memStatsMetrics[i].ID,
				fmt.Sprintf("%v", field.Interface()),
			)
		} else {
			slog.Info(
				"Telemetry",
				slog.String("message", "skip collecting metric"),
				slog.String("metric_id", memStatsMetrics[i].ID),
			)
		}
	}
}
