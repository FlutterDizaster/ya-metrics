package telemetry

import (
	"fmt"
	"log/slog"
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

type Worker interface {
	AddMetrics([]view.Metric)
}

type MetricsCollector struct {
	metricsList   []view.Metric
	metricsBuffer []view.Metric
	rnd           rand.Rand
}

func NewMetricCollector(metricsList []view.Metric) *MetricsCollector {
	slog.Debug("Creating metric collector")
	randSource := rand.NewSource(time.Now().UnixNano())
	return &MetricsCollector{
		metricsList: metricsList,
		rnd:         *rand.New(randSource),
	}
}

func (mc *MetricsCollector) CollectMetrics() []view.Metric {
	slog.Debug("Collecting metrics")
	// Очистка буфера
	mc.metricsBuffer = make([]view.Metric, 0)
	// Сохранение PollCounter
	mc.saveMetric(view.KindCounter, "PollCount", "1")

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
	slog.Debug("Adding metrics")
	return mc.metricsBuffer
}

func (mc *MetricsCollector) saveMetric(kind string, name string, value string) {
	// создание метрики
	metric, err := view.NewMetric(kind, name, value)
	if err != nil {
		slog.Error("%s metric not created: %s", name, err)
		return
	}
	// добавление метрики в буффер
	mc.metricsBuffer = append(mc.metricsBuffer, *metric)
}