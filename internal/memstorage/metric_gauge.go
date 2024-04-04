package memstorage

import "strconv"

type metricGauge struct {
	value float64
}

func (metric *metricGauge) UpdateValue(newValue string) error {
	fValue, err := strconv.ParseFloat(newValue, 64)
	if err != nil {
		return err
	}
	metric.value = fValue
	return nil
}

func (metric *metricGauge) GetValue() string {
	return strconv.FormatFloat(metric.value, 'f', -1, 64)
}

func (metric *metricGauge) Kind() string {
	return "gauge"
}
