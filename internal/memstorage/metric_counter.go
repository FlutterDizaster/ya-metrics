package memstorage

import "strconv"

type metricCounter struct {
	value int64
}

func (metric *metricCounter) UpdateValue(newValue string) error {
	iValue, err := strconv.ParseInt(newValue, 10, 64)
	if err != nil {
		return err
	}
	metric.value += iValue
	return nil
}

func (metric *metricCounter) GetValue() string {
	return strconv.FormatInt(metric.value, 10)
}

func (metric *metricCounter) Kind() string {
	return "counter"
}
