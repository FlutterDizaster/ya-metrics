package telemetry

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestTelemetry_collectMemStats(t *testing.T) {
	type test struct {
		name    string
		metrics []view.Metric
	}

	tests := []test{
		{
			name: "simple test",
			metrics: []view.Metric{
				{ID: "Alloc", MType: "gauge"},
				{ID: "BuckHashSys", MType: "gauge"},
				{ID: "Frees", MType: "gauge"},
				{ID: "GCCPUFraction", MType: "gauge"},
				{ID: "GCSys", MType: "gauge"},
				{ID: "HeapAlloc", MType: "gauge"},
				{ID: "HeapIdle", MType: "gauge"},
				{ID: "HeapInuse", MType: "gauge"},
				{ID: "HeapObjects", MType: "gauge"},
				{ID: "HeapReleased", MType: "gauge"},
				{ID: "HeapSys", MType: "gauge"},
				{ID: "LastGC", MType: "gauge"},
				{ID: "Lookups", MType: "gauge"},
				{ID: "MCacheInuse", MType: "gauge"},
				{ID: "MCacheSys", MType: "gauge"},
				{ID: "MSpanInuse", MType: "gauge"},
				{ID: "MSpanSys", MType: "gauge"},
				{ID: "Mallocs", MType: "gauge"},
				{ID: "NextGC", MType: "gauge"},
				{ID: "NumForcedGC", MType: "gauge"},
				{ID: "NumGC", MType: "gauge"},
				{ID: "OtherSys", MType: "gauge"},
				{ID: "PauseTotalNs", MType: "gauge"},
				{ID: "StackInuse", MType: "gauge"},
				{ID: "StackSys", MType: "gauge"},
				{ID: "Sys", MType: "gauge"},
				{ID: "TotalAlloc", MType: "gauge"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			telem := &Telemetry{}

			metrics := telem.collectMemStats()

			for _, metric := range tt.metrics {
				found := false

				for _, m := range metrics {
					if m.ID == metric.ID {
						found = true
						break
					}
				}
				assert.True(t, found)
			}
		})
	}
}

func TestTelemetry_collectPCStats(t *testing.T) {
	type test struct {
		name    string
		metrics []view.Metric
	}

	tests := []test{
		{
			name: "simple test",
			metrics: []view.Metric{
				{ID: "TotalMemory", MType: view.KindGauge},
				{ID: "FreeMemory", MType: view.KindGauge},
				{ID: "UsedMemory", MType: view.KindGauge},
			},
		},
	}

	cpuCount := runtime.NumCPU()

	for _, tt := range tests {
		for i := 0; i < cpuCount; i++ {
			tt.metrics = append(tt.metrics, view.Metric{
				ID:    fmt.Sprintf("CPUutilization%d", i+1),
				MType: view.KindGauge,
			})
		}

		t.Run(tt.name, func(t *testing.T) {
			telem := &Telemetry{}

			metrics := telem.collectPCStats()

			for _, metric := range tt.metrics {
				found := false

				for _, m := range metrics {
					if m.ID == metric.ID {
						found = true
						break
					}
				}
				assert.True(t, found)
			}
		})
	}
}
