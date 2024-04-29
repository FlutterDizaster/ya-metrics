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

type MetricStorage interface {
	AddMetric(view.Metric) error
}

type MetricsCollector struct {
	metricsList  []view.Metric
	pollInterval int
	storage      MetricStorage
	rnd          rand.Rand
}

func NewMetricCollector(
	storage MetricStorage,
	pollInterval int,
	metricsList []view.Metric,
) MetricsCollector {
	randSource := rand.NewSource(time.Now().UnixNano())
	return MetricsCollector{
		metricsList:  metricsList,
		pollInterval: pollInterval,
		storage:      storage,
		rnd:          *rand.New(randSource),
	}
}

func (mc *MetricsCollector) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			mc.CollectMetrics()
			return
		default:
			mc.CollectMetrics()
			time.Sleep(time.Duration(mc.pollInterval) * time.Second)
		}
	}
}

func (mc *MetricsCollector) CollectMetrics() {
	// Сохранение PollCounter
	mc.saveMetric(view.KindCounter, "PollCounter", "1")

	// Сохранение RandomValue
	rvalue := strconv.FormatFloat(mc.rnd.Float64(), 'f', -1, 64)
	mc.saveMetric(view.KindGauge, "RandomValue", rvalue)

	// Получение метрик MemStats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Сохранение метрик из списка
	for _, metric := range mc.metricsList {
		switch metric.Source {
		case view.MemStats:
			// Попытка получить поле metric.ID из memStats
			field := reflect.ValueOf(memStats).FieldByName(metric.ID)
			if field.IsValid() {
				mc.saveMetric(metric.MType, metric.ID, fmt.Sprintf("%v", field.Interface()))
			} else {
				slog.Info(fmt.Sprintf("Skip collection of metric \"%s\" because there is no metric with that name.", metric.ID))
			}
		case view.None:
			continue
		}
	}
}

func (mc *MetricsCollector) saveMetric(kind string, name string, value string) {
	// создание метрики
	metric, err := view.NewMetric(kind, name, value)
	if err != nil {
		slog.Error("%s metric not created: %s", name, err)
		return
	}

	// добавление метрики в storage
	err = mc.storage.AddMetric(*metric)
	if err != nil {
		slog.Error("%s metric not added to storage: %s", name, err)
	}
}
