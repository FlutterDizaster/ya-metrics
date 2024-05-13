package postgres

import (
	"context"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

// type DataProvider interface {
// 	AddGauge(view.Metric) (view.Metric, error)
// 	AddCounter(view.Metric) (view.Metric, error)
// 	GetMetric(kind string, name string) (view.Metric, error)
// 	ReadAllMetrics() []view.Metric
// }

// var _ DataProvider = &MetricStorage{}

type MetricStorage struct {
}

func New(_ string) *MetricStorage {
	return &MetricStorage{}
}

func (ms *MetricStorage) Start(_ context.Context) error {
	return nil
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
