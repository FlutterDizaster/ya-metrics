package storage

import (
	"errors"
	"strconv"
)

const (
	COUNTER = "counter"
	GAUGE   = "gauge"
)

type metric struct {
	metricType string
	value      interface{}
}

type MemStorage struct {
	metrics map[string]metric
}

func NewMemStorage() MemStorage {
	return MemStorage{
		metrics: make(map[string]metric),
	}
}

func (ms *MemStorage) AddMetricValue(mtype string, name string, value string) error {
	switch mtype {
	case COUNTER:
		return ms.addCounterValue(name, value)
	case GAUGE:
		return ms.addGaugeValue(name, value)
	default:
		return errors.New("wrong metrics type")
	}
}

func (ms *MemStorage) addCounterValue(name string, value string) error {
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
			metricType: COUNTER,
			value:      oldValue.value.(int64) + ivalue,
		}
	} else { //Если записи ещё нет
		ms.metrics[name] = metric{
			metricType: COUNTER,
			value:      ivalue,
		}
	}

	return nil
}

func (ms *MemStorage) addGaugeValue(name string, value string) error {
	//Проверка типа
	fvalue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}

	//Добавление/обновление метрики
	ms.metrics[name] = metric{
		metricType: GAUGE,
		value:      fvalue,
	}
	return nil
}
