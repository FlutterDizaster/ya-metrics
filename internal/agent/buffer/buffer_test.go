package buffer

import (
	"testing"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuffer_Put(t *testing.T) {
	type test struct {
		name      string
		testKind  string
		values    []view.Metric
		wantDelta int64
		wantValue float64
		wantErr   bool
	}

	tests := []test{
		{
			name:     "gauge test",
			testKind: view.KindGauge,
			values: []view.Metric{
				{
					ID:    view.KindGauge,
					MType: view.KindGauge,
					Value: func(i float64) *float64 { return &i }(45),
				},
				{
					ID:    view.KindGauge,
					MType: view.KindGauge,
					Value: func(i float64) *float64 { return &i }(54),
				},
			},
			wantValue: 54,
			wantErr:   false,
		},
		{
			name:     "counter test",
			testKind: view.KindCounter,
			values: []view.Metric{
				{
					ID:    view.KindCounter,
					MType: view.KindCounter,
					Delta: func(i int64) *int64 { return &i }(45),
				},
				{
					ID:    view.KindCounter,
					MType: view.KindCounter,
					Delta: func(i int64) *int64 { return &i }(54),
				},
			},
			wantDelta: 99,
			wantErr:   false,
		},
		{
			name:     "error test",
			testKind: "error",
			values:   []view.Metric{},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := New()

			if tt.wantErr {
				buffer.Close()
			}

			err := buffer.Put(tt.values)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Contains(t, buffer.metrics, tt.testKind)

			if tt.testKind == view.KindGauge {
				assert.InDelta(t, tt.wantValue, *buffer.metrics[view.KindGauge].Value, 0.001)
			} else if tt.testKind == view.KindCounter {
				assert.Equal(t, tt.wantDelta, *buffer.metrics[view.KindCounter].Delta)
			}
			buffer.Close()
		})
	}
}

func TestBuffer_Pull(t *testing.T) {
	type test struct {
		name    string
		values  []view.Metric
		wantErr bool
	}

	tests := []test{
		{
			name: "pull test",
			values: []view.Metric{
				{
					ID:    view.KindCounter,
					MType: view.KindCounter,
					Delta: func(i int64) *int64 { return &i }(45),
				},
				{
					ID:    view.KindGauge,
					MType: view.KindGauge,
					Value: func(i float64) *float64 { return &i }(54),
				},
			},
			wantErr: false,
		},
		{
			name:    "error test",
			values:  []view.Metric{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := New()

			if tt.wantErr {
				buffer.Close()
			} else {
				err := buffer.Put(tt.values)
				require.NoError(t, err)
			}

			metrics, err := buffer.Pull()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.values, metrics)

			buffer.Close()
		})
	}
}

func BenchmarkBuffer_Put(b *testing.B) {
	buffer := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := buffer.Put([]view.Metric{
			{
				ID:    view.KindCounter,
				MType: view.KindCounter,
				Delta: func(i int64) *int64 { return &i }(45),
			},
			{
				ID:    view.KindGauge,
				MType: view.KindGauge,
				Value: func(i float64) *float64 { return &i }(54),
			},
		})
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	buffer.Close()
	b.StartTimer()
}

// func BenchmarkBuffer_Pull(b *testing.B) {

// }
