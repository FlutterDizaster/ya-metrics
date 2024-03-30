package memstorage

import (
	"errors"
	"sync"
)

type metric interface {
	UpdateValue(newValue string) error
	GetValue() string
	Kind() string
}

type MetricStorage struct {
	metrics map[string]metric
	mtx     sync.Mutex
}

func NewMetricStorage() MetricStorage {
	return MetricStorage{
		metrics: make(map[string]metric),
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

func (ms *MetricStorage) ReadAllMetrics() []struct {
	Name  string
	Kind  string
	Value string
} {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	return ms.getAllMetrics()
}

func (ms *MetricStorage) PullAllMetrics() []struct {
	Name  string
	Kind  string
	Value string
} {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	metrics := ms.getAllMetrics()
	ms.metrics = make(map[string]metric)
	return metrics
}

func (ms *MetricStorage) getAllMetrics() []struct {
	Name  string
	Kind  string
	Value string
} {
	metrics := make([]struct {
		Name  string
		Kind  string
		Value string
	}, 0)

	for name, metric := range ms.metrics {
		metrics = append(metrics, struct {
			Name  string
			Kind  string
			Value string
		}{Name: name, Kind: metric.Kind(), Value: metric.GetValue()})
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
