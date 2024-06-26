package api

import (
	"errors"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

type MockMetricsStorage struct {
	pingErr error
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

func (m *MockMetricsStorage) AddMetrics(metrics ...view.Metric) ([]view.Metric, error) {
	m.content = append(m.content, metrics...)
	return metrics, nil
}

func (m *MockMetricsStorage) ReadAllMetrics() ([]view.Metric, error) {
	return m.content, nil
}

func (m *MockMetricsStorage) Ping() error {
	return m.pingErr
}
