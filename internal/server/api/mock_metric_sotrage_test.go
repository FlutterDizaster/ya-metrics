package api

import (
	"errors"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

type MockMetricsStorage struct {
	content []view.Metric
}

var _ MetricsStorage = &MockMetricsStorage{}

func (m *MockMetricsStorage) GetMetric(kind string, name string) (view.Metric, error) {
	for _, metric := range m.content {
		if metric.ID == name && metric.MType == kind {
			return metric, nil
		}
	}
	return view.Metric{}, errors.New("not found")
}

func (m *MockMetricsStorage) AddMetric(metric view.Metric) (view.Metric, error) {
	m.content = append(m.content, metric)
	return metric, nil
}

func (m *MockMetricsStorage) ReadAllMetrics() []view.Metric {
	return m.content
}
