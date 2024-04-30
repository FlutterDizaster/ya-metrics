package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/stretchr/testify/assert"
)

type TestMetric struct {
	Kind  string
	Value string
}

type TestStorage struct {
	Metrics []view.Metric
}

func (ts *TestStorage) AddMetrics(metrics []view.Metric) {
	ts.Metrics = append(ts.Metrics, metrics...)
}

func TestMetricsCollector_CollectMetrics(t *testing.T) {
	type fields struct {
		metricsList []view.Metric
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
				metricsList: []view.Metric{
					{
						ID:     "Alloc",
						MType:  "gauge",
						Source: view.MemStats,
					},
					{
						ID:     "Frees",
						MType:  "gauge",
						Source: view.MemStats,
					},
				},
				storage: TestStorage{make([]view.Metric, 0)},
			},
			want: 4,
		},
		{
			name: "wrong name test",
			fields: fields{
				metricsList: []view.Metric{
					{
						ID:     "wrong name",
						MType:  "count",
						Source: view.MemStats,
					},
				},
				storage: TestStorage{make([]view.Metric, 0)},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricCollector(&tt.fields.storage, 2, tt.fields.metricsList)

			mc.CollectMetrics()
			assert.Len(t, tt.fields.storage.Metrics, tt.want)
			//TODO: проверить, что метрики добавляются првильно
		})
	}
}

func TestMetricsCollector_Start(t *testing.T) {
	type fields struct {
		metricsList  []view.Metric
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
				metricsList: []view.Metric{
					{
						ID:     "Alloc",
						MType:  "gauge",
						Source: view.MemStats,
					},
					{
						ID:     "Frees",
						MType:  "gauge",
						Source: view.MemStats,
					},
				},
				storage:      TestStorage{make([]view.Metric, 0)},
				pollInterval: 1,
			},
			testDuration: 3,
			want:         2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := NewMetricCollector(
				&tt.fields.storage,
				tt.fields.pollInterval,
				tt.fields.metricsList,
			)

			ctx, cancle := context.WithCancel(context.Background())

			go func() {
				time.Sleep(time.Duration(tt.testDuration) * time.Second)
				cancle()
			}()

			mc.Start(ctx)
			//TODO: Проверки
			assert.Len(t, mc.metricsList, tt.want)
		})
	}
}
