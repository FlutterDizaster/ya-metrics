package newmemory

import (
	"errors"
	"sync"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

var (
	errNotFound  = errors.New("metric not found")
	errWrongType = errors.New("wrong metric type")
)

const (
	kindGauge   = "gauge"
	kindCounter = "counter"
)

// // type MSI interface {
// 	AddMetric(view.Metric) error
// 	GetMetric(kind string, name string) (view.Metric, error)
// 	ReadAllMetrics() []view.Metric
// 	PullAllMetrics() []view.Metric
// }

type MetricStorage struct {
	metrics map[string]view.Metric
	mtx     sync.Mutex
}

func NewMetricStorage() *MetricStorage {
	return &MetricStorage{
		metrics: make(map[string]view.Metric),
	}
}

func (ms *MetricStorage) AddMetric(metric view.Metric) (view.Metric, error) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	switch metric.MType {
	case kindGauge:
		return ms.addGauge(metric)
	case kindCounter:
		return ms.addCounter(metric)
	default:
		return metric, errWrongType
	}
}

func (ms *MetricStorage) GetMetric(kind string, name string) (view.Metric, error) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	metric, ok := ms.metrics[name]
	if !ok {
		return view.Metric{}, errNotFound
	}

	if metric.MType != kind {
		return view.Metric{}, errWrongType
	}

	return metric, nil
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

	ms.metrics = make(map[string]view.Metric)

	return metrics
}

func (ms *MetricStorage) addCounter(metric view.Metric) (view.Metric, error) {
	oldMetric, ok := ms.metrics[metric.ID]
	if !ok {
		ms.metrics[metric.ID] = metric
		return metric, nil
	}

	if oldMetric.MType != metric.MType {
		return metric, errWrongType
	}

	delta := *oldMetric.Delta + *metric.Delta
	metric.Delta = &delta

	ms.metrics[metric.ID] = metric

	return metric, nil
}

func (ms *MetricStorage) addGauge(metric view.Metric) (view.Metric, error) {
	oldMetric, ok := ms.metrics[metric.ID]
	if !ok {
		ms.metrics[metric.ID] = metric
		return metric, nil
	}

	if oldMetric.MType != metric.MType {
		return metric, errWrongType
	}

	ms.metrics[metric.ID] = metric

	return metric, nil
}

func (ms *MetricStorage) getAllMetrics() []view.Metric {
	metrics := make([]view.Metric, len(ms.metrics))
	iter := 0

	for _, metric := range ms.metrics {
		metrics[iter] = metric
		iter++
	}

	return metrics
}
