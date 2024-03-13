package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type TestMetric struct {
	Name  string
	Kind  string
	Value string
}

type TestStorage struct {
	Metrics []TestMetric
}

func (ts *TestStorage) AddMetricValue(kind string, name string, value string) error {
	ts.Metrics = append(ts.Metrics, TestMetric{
		Name:  name,
		Kind:  kind,
		Value: value,
	})

	return nil
}

func TestMetricsCollector_CollectMetrics(t *testing.T) {
	type fields struct {
		metricsList []Metric
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
				metricsList: []Metric{
					{
						Name: "Alloc",
						Kind: "gauge",
					},
					{
						Name: "Frees",
						Kind: "gauge",
					},
				},
				storage: TestStorage{make([]TestMetric, 0)},
			},
			want: 2,
		},
		{
			name: "wrong name test",
			fields: fields{
				metricsList: []Metric{
					{
						Name: "wrong name",
						Kind: "count",
					},
				},
				storage: TestStorage{make([]TestMetric, 0)},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MetricsCollector{
				metricsList: tt.fields.metricsList,
				storage:     &tt.fields.storage,
			}
			mc.CollectMetrics()
			//Проверяем добавились ли метрики
			require.Equal(t, tt.want, len(tt.fields.storage.Metrics))
			for _, metric := range tt.fields.storage.Metrics {
				t.Logf("metric \"%s\" type of \"%s\" have value %s\n", metric.Name, metric.Kind, metric.Value)
			}
			//TODO проверить, что метрики добавляются првильно
		})
	}
}

func TestMetricsCollector_Start(t *testing.T) {
	type fields struct {
		metricsList  []Metric
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
				metricsList: []Metric{
					{
						Name: "Alloc",
						Kind: "gauge",
					},
					{
						Name: "Frees",
						Kind: "gauge",
					},
				},
				storage:      TestStorage{make([]TestMetric, 0)},
				pollInterval: 1,
			},
			testDuration: 3,
			want:         8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricCollector(&tt.fields.storage, tt.fields.pollInterval, tt.fields.metricsList)

			ctx, cancle := context.WithCancel(context.Background())

			go func() {
				time.Sleep(time.Duration(tt.testDuration) * time.Second)
				cancle()
			}()

			mc.Start(ctx)
			// TODO: Проверки
			require.Equal(t, tt.want, len(tt.fields.storage.Metrics))
			for _, metric := range tt.fields.storage.Metrics {
				t.Logf("metric \"%s\" type of \"%s\" have value %s\n", metric.Name, metric.Kind, metric.Value)
			}
		})
	}
}
