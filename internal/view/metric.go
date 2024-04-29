package view

import (
	"errors"
	"strconv"
)

type MetricSource int

const (
	None MetricSource = iota
	MemStats
)

const (
	KindGauge   = "gauge"
	KindCounter = "counter"
)

type Metric struct {
	ID     string       `json:"name"`
	MType  string       `json:"type"`
	Delta  *int64       `json:"delta,omitempty"`
	Value  *float64     `json:"value,omitempty"`
	Source MetricSource `json:"-"`
}

func NewMetric(kind string, name string, value string) (*Metric, error) {
	metric := &Metric{
		ID:    name,
		MType: kind,
	}
	switch kind {
	case KindGauge:
		fvalue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		metric.Value = &fvalue
	case KindCounter:
		delta, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		metric.Delta = &delta
	default:
		return nil, errors.New("wrong metric type")
	}
	return metric, nil
}

func (m *Metric) StringValue() string {
	switch m.MType {
	case "gauge":
		return strconv.FormatFloat(*m.Value, 'f', -1, 64)
	case "counter":
		return strconv.FormatInt(*m.Delta, 10)
	}
	return ""
}
