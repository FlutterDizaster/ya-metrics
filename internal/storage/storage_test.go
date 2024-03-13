package storage

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetricStorage_addGaugeValue(t *testing.T) {
	type values struct {
		name      string
		value     string
		wantErr   bool
		wantValue float64
	}
	tests := []struct {
		name   string
		values []values
	}{
		{
			name: "simple test",
			values: []values{
				{
					name:      "testvalue",
					value:     "0.001",
					wantErr:   false,
					wantValue: 0.001,
				},
				{
					name:      "testvalue",
					value:     "0.555",
					wantErr:   false,
					wantValue: 0.555,
				},
			},
		},
		{
			name: "int test",
			values: []values{
				{
					name:      "testvalue",
					value:     "55",
					wantErr:   false,
					wantValue: 55,
				},
			},
		},
		{
			name: "error test",
			values: []values{
				{
					name:    "testvalue",
					value:   "test",
					wantErr: true,
				},
			},
		},
		{
			name: "semi error test",
			values: []values{
				{
					name:      "testvalue",
					value:     "0.555",
					wantErr:   false,
					wantValue: 0.555,
				},
				{
					name:      "testvalue",
					value:     "test",
					wantErr:   true,
					wantValue: 0.555,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := NewMetricStorage()
			for _, value := range tt.values {
				err := ms.addGaugeValue(value.name, value.value)
				// Error check
				require.Truef(t, (err != nil) == value.wantErr, "MetricStorage.addGaugeValue() error = %v, wantErr %v", err, value.wantErr)

				//Content check
				mapValue := ms.metrics[value.name]
				if !value.wantErr {
					require.Equal(t, value.wantValue, mapValue.value)
				}
			}
		})
	}
}

func TestMetricStorage_addCounterValue(t *testing.T) {
	type values struct {
		name      string
		value     string
		wantErr   bool
		wantValue int64
	}
	tests := []struct {
		name   string
		values []values
	}{
		{
			name: "simple test",
			values: []values{
				{
					name:      "testvalue",
					value:     "1",
					wantErr:   false,
					wantValue: 1,
				},
				{
					name:      "testvalue",
					value:     "555",
					wantErr:   false,
					wantValue: 556,
				},
			},
		},
		{
			name: "float test",
			values: []values{
				{
					name:    "testvalue",
					value:   "5.5",
					wantErr: true,
				},
			},
		},
		{
			name: "error test",
			values: []values{
				{
					name:    "testvalue",
					value:   "test",
					wantErr: true,
				},
			},
		},
		{
			name: "semi error test",
			values: []values{
				{
					name:      "testvalue",
					value:     "555",
					wantErr:   false,
					wantValue: 555,
				},
				{
					name:      "testvalue",
					value:     "test",
					wantErr:   true,
					wantValue: 555,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := NewMetricStorage()
			for _, value := range tt.values {
				err := ms.addCounterValue(value.name, value.value)
				// Error check
				require.Truef(t, (err != nil) == value.wantErr, "MetricStorage.addCounterValue() error = %v, wantErr %v", err, value.wantErr)

				//Content check
				mapValue := ms.metrics[value.name]
				if !value.wantErr {
					require.Equal(t, value.wantValue, mapValue.value)
				}
			}
		})
	}
}

func TestMetricStorage_AddMetricValue(t *testing.T) {
	type values struct {
		name      string
		kind      string
		value     string
		wantErr   bool
		wantValue interface{}
	}
	tests := []struct {
		name   string
		values []values
	}{
		{
			name: "simple float test",
			values: []values{
				{
					name:      "testvalue",
					kind:      KindGauge,
					value:     "0.001",
					wantErr:   false,
					wantValue: 0.001,
				},
				{
					name:      "testvalue",
					kind:      KindGauge,
					value:     "0.555",
					wantErr:   false,
					wantValue: 0.555,
				},
			},
		},
		{
			name: "float with int test",
			values: []values{
				{
					name:      "testvalue",
					kind:      KindGauge,
					value:     "55",
					wantErr:   false,
					wantValue: 55,
				},
			},
		},
		{
			name: "float error test",
			values: []values{
				{
					name:    "testvalue",
					kind:    KindGauge,
					value:   "test",
					wantErr: true,
				},
			},
		},
		{
			name: "float semi error test",
			values: []values{
				{
					name:      "testvalue",
					kind:      KindGauge,
					value:     "0.555",
					wantErr:   false,
					wantValue: 0.555,
				},
				{
					name:      "testvalue",
					kind:      KindGauge,
					value:     "test",
					wantErr:   true,
					wantValue: 0.555,
				},
			},
		},
		{
			name: "simple int test",
			values: []values{
				{
					name:      "testvalue",
					kind:      KindCounter,
					value:     "1",
					wantErr:   false,
					wantValue: 1,
				},
				{
					name:      "testvalue",
					kind:      KindCounter,
					value:     "555",
					wantErr:   false,
					wantValue: 556,
				},
			},
		},
		{
			name: "int with float test",
			values: []values{
				{
					name:    "testvalue",
					kind:    KindCounter,
					value:   "5.5",
					wantErr: true,
				},
			},
		},
		{
			name: "int error test",
			values: []values{
				{
					name:    "testvalue",
					kind:    KindCounter,
					value:   "test",
					wantErr: true,
				},
			},
		},
		{
			name: "int semi error test",
			values: []values{
				{
					name:      "testvalue",
					kind:      KindCounter,
					value:     "555",
					wantErr:   false,
					wantValue: 555,
				},
				{
					name:      "testvalue",
					kind:      KindCounter,
					value:     "test",
					wantErr:   true,
					wantValue: 555,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := NewMetricStorage()
			for _, value := range tt.values {
				err := ms.AddMetricValue(value.kind, value.name, value.value)
				// Error check
				require.Truef(t, (err != nil) == value.wantErr, "MetricStorage.AddMetricValue() error = %v, wantErr %v", err, value.wantErr)

				//Content check
				mapValue := ms.metrics[value.name]
				if !value.wantErr {
					require.EqualValues(t, value.wantValue, mapValue.value)
				}
			}
		})
	}
}

func TestMetricStorage_GetAll(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]metric
		want   []struct {
			Name  string
			Kind  string
			Value string
		}
	}{
		{
			name: "simple test",
			values: map[string]metric{
				"test1": {KindCounter, int64(1)},
				"test2": {KindGauge, float64(1)},
				"test3": {KindGauge, 0.1},
			},
			want: []struct {
				Name  string
				Kind  string
				Value string
			}{
				{"test1", KindCounter, "1"},
				{"test2", KindGauge, "1"},
				{"test3", KindGauge, "0.1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MetricStorage{
				metrics: tt.values,
				mtx:     sync.Mutex{},
			}
			got := ms.GetAll()
			fmt.Println(got)
			require.ElementsMatchf(t, tt.want, got, "MetricStorage.GetAll() = %v, want %v", got, tt.want)
		})
	}
}
