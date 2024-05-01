package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

func (w *Worker) startCollecting(ctx context.Context) {
	slog.Debug("Start collecting metrics")
	ticker := time.NewTicker(time.Duration(w.pollInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			w.collect()
			return
		default:
			w.collect()
		}
		<-ticker.C
	}
}

func (w *Worker) collect() {
	metrics := w.collector.CollectMetrics()

	w.cond.L.Lock()
	defer w.cond.L.Unlock()

	for _, metric := range metrics {
		switch metric.MType {
		case view.KindCounter:
			old, ok := w.buffer[metric.ID]
			if !ok {
				w.buffer[metric.ID] = metric
				continue
			}
			delta := *metric.Delta + *old.Delta
			metric.Delta = &delta
			w.buffer[metric.ID] = metric
		case view.KindGauge:
			w.buffer[metric.ID] = metric
		}
	}

	w.cond.Broadcast()
}
