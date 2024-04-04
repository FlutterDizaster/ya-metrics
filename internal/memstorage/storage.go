package memstorage

import (
	"errors"
	"sync"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

type Metric interface {
	UpdateValue(newValue string) error
	GetValue() string
	Kind() string
}

type MetricStorage struct {
	metrics map[string]Metric
	mtx     sync.Mutex
}

func NewMetricStorage() MetricStorage {
	return MetricStorage{
		metrics: make(map[string]Metric),
	}
}

func (ms *MetricStorage) AddMetricValue(kind string, name string, value string) error {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	metric, ok := ms.metrics[name]

	if !ok {
		switch kind {
		case "gauge":
			metric = &metricGauge{}
		case "counter":
			metric = &metricCounter{}
		}
	}

	if metric.Kind() != kind {
		return errors.New("missmatch metric types")
	}

	err := metric.UpdateValue(value)
	if err != nil {
		return err
	}

	ms.metrics[name] = metric

	return nil
}

func (ms *MetricStorage) ReadAllMetrics() []view.Metric {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	return ms.getAllMetrics()
}

func (ms *MetricStorage) PullAllMetrics() []view.Metric {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	metrics := ms.getAllMetrics()
	ms.metrics = make(map[string]Metric)
	return metrics
}

func (ms *MetricStorage) getAllMetrics() []view.Metric {
	metrics := make([]view.Metric, 0)

	for name, metric := range ms.metrics {
		metrics = append(
			metrics,
			view.Metric{Name: name, Kind: metric.Kind(), Value: metric.GetValue()},
		)
	}

	return metrics
}

func (ms *MetricStorage) GetMetricValue(kind string, name string) (string, error) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	value, ok := ms.metrics[name]
	if !ok {
		return "", errors.New("metric not found")
	}

	if kind != value.Kind() {
		return "", errors.New("missmatch value types")
	}

	return value.GetValue(), nil
}
