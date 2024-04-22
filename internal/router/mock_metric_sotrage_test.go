package router

import (
	"errors"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

type MockMetricsStorage struct {
	content []view.Metric
}

func (m *MockMetricsStorage) AddMetricValue(kind string, name string, value string) error {
	m.content = append(m.content, view.Metric{
		Name:  name,
		Kind:  kind,
		Value: value,
	})
	return nil
}

func (m *MockMetricsStorage) GetMetricValue(_ string, name string) (string, error) {
	var value string
	var err error

	for _, v := range m.content {
		if v.Name == name {
			value = v.Value
		}
	}
	if value == "" {
		err = errors.New("Not found")
	}

	return value, err
}

func (m *MockMetricsStorage) ReadAllMetrics() []view.Metric {
	return m.content
}
