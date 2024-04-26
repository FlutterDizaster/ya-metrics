package router

import (
	"errors"
	"strconv"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

type MockMetricsStorage struct {
	content []view.Metric
}

func (m *MockMetricsStorage) AddMetricValue(kind string, name string, value string) error {
	metric := view.Metric{
		ID:    name,
		MType: kind,
	}
	switch kind {
	case gauge:
		fvalue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		metric.Value = &fvalue
	case counter:
		delta, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		metric.Delta = &delta
	}
	m.content = append(m.content, metric)
	return nil
}

func (m *MockMetricsStorage) GetMetricValue(_ string, name string) (string, error) {
	var value string
	var err error

	for _, v := range m.content {
		if v.ID == name {
			// value = v.Value
			switch v.MType {
			case gauge:
				value = strconv.FormatFloat(*v.Value, 'f', -1, 64)
			case counter:
				value = strconv.FormatInt(*v.Delta, 10)
			}
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
