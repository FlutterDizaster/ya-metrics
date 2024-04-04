package handlers

import "github.com/FlutterDizaster/ya-metrics/internal/view"

type AddMetricStorage interface {
	AddMetricValue(kind string, name string, value string) error
}

type GetMetricStorage interface {
	GetMetricValue(kind string, name string) (string, error)
}

type GetAllMetricsStorage interface {
	ReadAllMetrics() []view.Metric
}
