package telemetry

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"time"
)

type ValueKind string

const (
	KindGauge   ValueKind = "float64"
	KindCounter ValueKind = "int64"
)

type MetricStorage interface {
	AddMetricValue(kind string, name string, value string) error
}

type Metric struct {
	Name string
	Kind ValueKind
}

type MetricsCollector struct {
	metricsList  []Metric
	pollInterval int
	storage      MetricStorage
	randSource   rand.Source
}

func NewMetricCollector(storage MetricStorage, pollInterval int, metricsList []Metric) MetricsCollector {
	return MetricsCollector{
		metricsList:  metricsList,
		pollInterval: pollInterval,
		storage:      storage,
		randSource:   rand.NewSource(time.Now().UnixNano()),
	}
}

func (mc *MetricsCollector) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			mc.CollectMetrics()
			return
		default:
			mc.CollectMetrics()
			time.Sleep(time.Duration(mc.pollInterval) * time.Second)
		}
	}
}

func (mc *MetricsCollector) CollectMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	for _, metric := range mc.metricsList {
		if metric.Name == "PollCount" {
			// Add pollCount metric to storage
			mc.storage.AddMetricValue(metric.Name, string(metric.Kind), strconv.Itoa(1))
		} else if metric.Name == "RandomValue" {
			//TODO: RandomValue logic
		} else {
			//try to get field with name metric.Name from memStats
			field := reflect.ValueOf(memStats).FieldByName(metric.Name)
			if field.IsValid() {
				mc.storage.AddMetricValue(string(metric.Kind), metric.Name, fmt.Sprintf("%v", field.Interface()))
			} else {
				log.Printf("Skip collection of metric \"%s\" because there is no metric with that name.", metric.Name)
			}
		}
	}
}
