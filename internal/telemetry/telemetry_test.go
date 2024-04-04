package telemetry_test

import (
	"context"
	"testing"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/telemetry"
	"github.com/stretchr/testify/assert"
)

type TestMetric struct {
	Kind  string
	Value string
}

type TestStorage struct {
	Metrics map[string]TestMetric
}

func (ts *TestStorage) AddMetricValue(kind string, name string, value string) error {
	ts.Metrics[name] = TestMetric{
		Kind:  kind,
		Value: value,
	}

	return nil
}

func TestMetricsCollector_CollectMetrics(t *testing.T) {
	type fields struct {
		metricsList []telemetry.Metric
		storage     TestStorage
	}

	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "simple test",
			fields: fields{
				metricsList: []telemetry.Metric{
					{
						Name: "Alloc",
						Kind: "gauge",
					},
					{
						Name: "Frees",
						Kind: "gauge",
					},
				},
				storage: TestStorage{make(map[string]TestMetric)},
			},
			want: 2,
		},
		{
			name: "wrong name test",
			fields: fields{
				metricsList: []telemetry.Metric{
					{
						Name: "wrong name",
						Kind: "count",
					},
				},
				storage: TestStorage{make(map[string]TestMetric)},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &telemetry.MetricsCollector{
				MetricsList: tt.fields.metricsList,
				Storage:     &tt.fields.storage,
			}
			mc.CollectMetrics()
			assert.Len(t, tt.fields.storage.Metrics, tt.want)
			//TODO: проверить, что метрики добавляются првильно
		})
	}
}

func TestMetricsCollector_Start(t *testing.T) {
	type fields struct {
		metricsList  []telemetry.Metric
		storage      TestStorage
		pollInterval int
	}
	tests := []struct {
		name         string
		fields       fields
		testDuration int
		want         int
	}{
		{
			name: "simple test",
			fields: fields{
				metricsList: []telemetry.Metric{
					{
						Name: "Alloc",
						Kind: "gauge",
					},
					{
						Name: "Frees",
						Kind: "gauge",
					},
				},
				storage:      TestStorage{make(map[string]TestMetric)},
				pollInterval: 1,
			},
			testDuration: 3,
			want:         2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := telemetry.NewMetricCollector(&tt.fields.storage, tt.fields.pollInterval, tt.fields.metricsList)

			ctx, cancle := context.WithCancel(context.Background())

			go func() {
				time.Sleep(time.Duration(tt.testDuration) * time.Second)
				cancle()
			}()

			mc.Start(ctx)
			//TODO: Проверки
			assert.Len(t, mc.MetricsList, tt.want)
		})
	}
}
