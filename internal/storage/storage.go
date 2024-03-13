package storage

import (
	"errors"
	"strconv"
	"sync"
)

const (
	KindGauge   = "gauge"
	KindCounter = "counter"
)

type metric struct {
	kind  string
	value interface{}
}

type MetricStorage struct {
	metrics map[string]metric
	mtx     sync.Mutex
}

func NewMetricStorage() MetricStorage {
	return MetricStorage{
		metrics: make(map[string]metric),
		mtx:     sync.Mutex{},
	}
}

func (ms *MetricStorage) GetAll() []struct {
	Name  string
	Kind  string
	Value string
} {
	result := make([]struct {
		Name  string
		Kind  string
		Value string
	}, 0)

	//Locking mtx
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	for name, metric := range ms.metrics {
		//parsing value
		var value string
		switch metric.kind {
		case KindGauge:
			rawValue := metric.value.(float64)
			value = strconv.FormatFloat(rawValue, 'f', -1, 64)
		case KindCounter:
			rawValue := metric.value.(int64)
			value = strconv.FormatInt(rawValue, 10)
		}

		//adding to result slice
		result = append(result, struct {
			Name  string
			Kind  string
			Value string
		}{
			Name:  name,
			Kind:  metric.kind,
			Value: value,
		})
	}

	//cleanup storage
	ms.metrics = make(map[string]metric)

	return result
}

func (ms *MetricStorage) AddMetricValue(kind string, name string, value string) error {
	//Locking mtx
	ms.mtx.Lock()
	defer ms.mtx.Unlock()

	switch kind {
	case KindCounter:
		return ms.addCounterValue(name, value)
	case KindGauge:
		return ms.addGaugeValue(name, value)
	default:
		return errors.New("wrong metrics type")
	}
}

func (ms *MetricStorage) addCounterValue(name string, value string) error {
	//Проверка типа
	ivalue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}

	//Проверка существует ли метрика
	oldValue, ok := ms.metrics[name]

	//Добавление метрики
	if ok { //Если запись уже имеется
		ms.metrics[name] = metric{
			kind:  KindCounter,
			value: oldValue.value.(int64) + ivalue,
		}
	} else { //Если записи ещё нет
		ms.metrics[name] = metric{
			kind:  KindCounter,
			value: ivalue,
		}
	}

	return nil
}

func (ms *MetricStorage) addGaugeValue(name string, value string) error {
	//Проверка типа
	fvalue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}

	//Добавление/обновление метрики
	ms.metrics[name] = metric{
		kind:  KindGauge,
		value: fvalue,
	}
	return nil
}
