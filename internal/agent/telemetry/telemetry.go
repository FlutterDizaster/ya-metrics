package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

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
	// Создание буфера
	metrics := make([]view.Metric, 0)

	// Получение PollCount
	metric, err := view.NewMetric(view.KindCounter, "PollCount", "1")
	if err != nil {
		slog.Error(
			"metric not created error",
			slog.String("metric", "PollCount"),
			slog.Any("error", err),
		)
	} else {
		metrics = append(metrics, *metric)
	}

	// Получение RandomValue
	rvalue := strconv.FormatFloat(t.rnd.Float64(), 'f', -1, 64)
	metric, err = view.NewMetric(view.KindGauge, "RandomValue", rvalue)
	if err != nil {
		slog.Error(
			"metric not created error",
			slog.String("metric", "RandomValue"),
			slog.Any("error", err),
		)
	} else {
		metrics = append(metrics, *metric)
	}

	wg := sync.WaitGroup{}

	// Получение метрик MemStats
	var memStats []view.Metric
	wg.Add(1)
	go func() {
		memStats = t.collectMemStats()
		wg.Done()
	}()

	// Получение метрик PCMetrics
	var pcStats []view.Metric
	wg.Add(1)
	go func() {
		pcStats = t.collectPCStats()
		wg.Done()
	}()

	// Ожидание конца сбора метрик
	wg.Wait()

	metrics = append(metrics, pcStats...)
	metrics = append(metrics, memStats...)

	// Отправка метрик в буффер приложения
	err = t.buf.Put(metrics)
	if err != nil {
		slog.Error("Telemetry", "error", err)
		return
	}

	slog.Debug("Telemetry", slog.String("status", "collected"))
}

func (t *Telemetry) collectPCStats() view.Metrics {
	metrics := make([]view.Metric, 0)

	// Получение статистики памяти
	vmStats, err := mem.VirtualMemory()
	if err != nil {
		slog.Error("error reading VirtualMempry stats", "error", err)
	}
	// Сохранение TotalMemory
	metric, err := view.NewMetric(
		view.KindGauge,
		"TotalMemory",
		strconv.FormatUint(vmStats.Total, 10),
	)
	if err != nil {
		slog.Error(
			"metric not created error",
			slog.String("metric", "PollCount"),
			slog.Any("error", err),
		)
	} else {
		metrics = append(metrics, *metric)
	}

	// Сохранение FreeMemory
	metric, err = view.NewMetric(view.KindGauge, "FreeMemory", strconv.FormatUint(vmStats.Free, 10))
	if err != nil {
		slog.Error(
			"metric not created error",
			slog.String("metric", "PollCount"),
			slog.Any("error", err),
		)
	} else {
		metrics = append(metrics, *metric)
	}

	utilization, err := cpu.Percent(0, true)
	if err != nil {
		slog.Error("error reading CPU stats", "error", err)
	}

	for i := range utilization {
		name := fmt.Sprintf("CPUutilization%d", i+1)
		metric, err = view.NewMetric(
			view.KindGauge,
			name,
			strconv.FormatFloat(utilization[i], 'f', -1, 64),
		)
		if err != nil {
			slog.Error(
				"metric not created error",
				slog.String("metric", name),
				slog.Any("error", err),
			)
		} else {
			metrics = append(metrics, *metric)
		}
	}

	return metrics
}

func (t *Telemetry) collectMemStats() view.Metrics {
	// Список интересующих метрик
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

	metrics := make([]view.Metric, 0, len(memStatsMetrics))

	// Получение метрик MemStats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Парсинг метрик
	for i := range memStatsMetrics {
		field := reflect.ValueOf(memStats).FieldByName((memStatsMetrics[i].ID))

		// Если такое поле существует, получаем его данные
		if field.IsValid() {
			// Создание метрики
			metric, err := view.NewMetric(
				memStatsMetrics[i].MType,
				memStatsMetrics[i].ID,
				fmt.Sprintf("%v", field.Interface()),
			)
			// если ошибка, то логирование
			if err != nil {
				slog.Error(
					"metric not created error",
					slog.String("metric", memStatsMetrics[i].ID),
					slog.Any("error", err),
				)
			} else {
				// Есди ощибки нет, то добавляем метрику к слайсу
				metrics = append(metrics, *metric)
			}
		} else {
			// Если поля нет, логируем
			slog.Info(
				"Telemetry",
				slog.String("message", "skip collecting metric"),
				slog.String("metric_id", memStatsMetrics[i].ID),
			)
		}
	}

	return metrics
}
