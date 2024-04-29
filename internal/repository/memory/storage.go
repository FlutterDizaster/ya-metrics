package memstorage

import (
	"errors"
	"log/slog"
	"strconv"
	"sync"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

const (
	metricKindGauge   = "gauge"
	metricKindCounter = "counter"
)

var (
	errWrongType  = errors.New("missmatch metric types")
	errNotFound   = errors.New("metric not found")
	errParseError = errors.New("unexpected parsing error")
)

type Metric interface {
	UpdateValue(newValue string) error
	Kind() string
	GetValue() string
	RawValue() interface{}
}

// TODO: Переписать Metric Storage, чтобы он использовал view.Metric, дабы избежать маппинга.
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
		case metricKindGauge:
			metric = &metricGauge{}
		case metricKindCounter:
			metric = &metricCounter{}
		}
	}

	if metric.Kind() != kind {
		return errWrongType
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
		newMetric := view.Metric{ID: name, MType: metric.Kind()}

		switch metric.Kind() {
		case metricKindGauge:
			fvalue, err := strconv.ParseFloat(metric.GetValue(), 64)
			if err != nil {
				slog.Error("error parsing metric %s value: %s", name, metric.GetValue())
				continue
			}
			newMetric.Value = &fvalue
		case metricKindCounter:
			delta, err := strconv.ParseInt(metric.GetValue(), 10, 64)
			if err != nil {
				slog.Error("error parsing metric %s value: %s", name, metric.GetValue())
				continue
			}
			newMetric.Delta = &delta
		}

		metrics = append(
			metrics,
			newMetric,
		)
	}

	return metrics
}

func (ms *MetricStorage) GetMetric(kind string, name string) (*view.Metric, error) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	smetric, ok := ms.metrics[name]
	if !ok {
		return &view.Metric{}, errNotFound
	}

	if smetric.Kind() != kind {
		return &view.Metric{}, errWrongType
	}

	nmetric := &view.Metric{
		ID:    name,
		MType: kind,
	}

	switch kind {
	case metricKindCounter:
		delta, err := smetric.RawValue().(int64)
		if !err {
			return &view.Metric{}, errParseError
		}
		nmetric.Delta = &delta
	case metricKindGauge:
		value, err := smetric.RawValue().(float64)
		if !err {
			return &view.Metric{}, errParseError
		}
		nmetric.Value = &value
	}

	return nmetric, nil
}

func (ms *MetricStorage) GetMetricValue(kind string, name string) (string, error) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	value, ok := ms.metrics[name]
	if !ok {
		return "", errNotFound
	}

	if kind != value.Kind() {
		return "", errWrongType
	}

	return value.GetValue(), nil
}
