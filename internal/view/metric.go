package view

import (
	"errors"
	"strconv"
)

const (
	KindGauge   = "gauge"   // Тип метрики gauge, значение метрики - float64
	KindCounter = "counter" // Тип метрики counter, значение метрики - int64
)

// Alias к срезу метрик.
//
//easyjson:json
type Metrics []Metric

// Metric - структура описывающая метрику.
// Может иметь тип gauge или counter.
//
// @swagger:model
//
//go:generate easyjson -all metric.go
type Metric struct {
	// Metric ID
	// Required: true
	ID string `json:"id"`
	// Metric Type
	// Possible values: gauge, counter
	// Required: true
	MType string `json:"type"`
	// Counter value
	// Required: false
	Delta *int64 `json:"delta,omitempty"`
	// Gauge value
	// Required: false
	Value *float64 `json:"value,omitempty"`
}

// Функция для создания метрики.
// kind - тип метрики KindGauge или KindCounter.
// name - имя метрики. Может быть любым.
// value - значение метрики. Должно быть текстовой репрезентацией целочисленного типа для метрик KindCounter или
// дробной для метрик KindGauge.
// При передаче некорректных значений возвращает ошибку.
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

// StringValue возвращает строковое представление значения метрики.
func (m *Metric) StringValue() string {
	switch m.MType {
	case "gauge":
		return strconv.FormatFloat(*m.Value, 'f', -1, 64)
	case "counter":
		return strconv.FormatInt(*m.Delta, 10)
	}
	return ""
}
