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
	KindGauge   ValueKind = "gauge"
	KindCounter ValueKind = "counter"
)

type MetricStorage interface {
	AddMetricValue(kind string, name string, value string) error
}

type Metric struct {
	Name string
	Kind ValueKind
}

type MetricsCollector struct {
	MetricsList  []Metric
	pollInterval int
	Storage      MetricStorage
	rnd          rand.Rand
}

func NewMetricCollector(
	storage MetricStorage,
	pollInterval int,
	metricsList []Metric,
) MetricsCollector {
	randSource := rand.NewSource(time.Now().UnixNano())
	return MetricsCollector{
		MetricsList:  metricsList,
		pollInterval: pollInterval,
		Storage:      storage,
		rnd:          *rand.New(randSource),
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

	// Add pollCount metric to storage
	err := mc.Storage.AddMetricValue(string(KindCounter), "PollCount", strconv.Itoa(1))
	if err != nil {
		log.Fatalf("Error %s while adding metric %s", err, "PollCount")
	}

	// Add random metric to storage
	randomValue := mc.rnd.Float64()
	err = mc.Storage.AddMetricValue(
		string(KindGauge),
		"RandomValue",
		strconv.FormatFloat(randomValue, 'f', -1, 64),
	)
	if err != nil {
		log.Fatalf("Error %s while adding metric %s", err, "RandomValue")
	}

	for _, metric := range mc.MetricsList {
		// try to get field with name metric.Name from memStats
		field := reflect.ValueOf(memStats).FieldByName(metric.Name)
		if field.IsValid() {
			err = mc.Storage.AddMetricValue(
				string(metric.Kind),
				metric.Name,
				fmt.Sprintf("%v", field.Interface()),
			)
			if err != nil {
				log.Fatalf("Error %s while adding metric %s", err, metric.Name)
			}
		} else {
			log.Printf("Skip collection of metric \"%s\" because there is no metric with that name.", metric.Name)
		}
	}
}
