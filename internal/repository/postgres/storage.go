package postgres

import "github.com/FlutterDizaster/ya-metrics/internal/view"

type Settings struct {
}

type MetricStorage struct {
}

func NewMetricStorage(_ *Settings) *MetricStorage {
	return &MetricStorage{}
}

func (ms *MetricStorage) AddMetric(_ view.Metric) (view.Metric, error) {
	return view.Metric{}, nil
}

func (ms *MetricStorage) GetMetric(_ string, _ string) (view.Metric, error) {
	return view.Metric{}, nil
}

func (ms *MetricStorage) ReadAllMetrics() []view.Metric {
	return make([]view.Metric, 0)
}
