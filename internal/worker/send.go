package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

func (w *Worker) startSending(ctx context.Context) {
	slog.Debug("Start sending metrics")
	ticker := time.NewTicker(time.Duration(w.reportInterval) * time.Second)
	// Первая отправка метрик
	w.send()
	for {
		select {
		case <-ctx.Done():
			w.send()
			// Выходим из функции
			return
		case <-ticker.C:
			w.send()
		}
	}
}

func (w *Worker) send() {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	// Ждем добавления метрик
	slog.Debug("Waiting metrics")

	w.cond.Wait()

	w.wg.Add(1)
	go func() {
		w.sender.SendMetrics(w.pullBuffer())
		w.wg.Done()
	}()
}

func (w *Worker) pullBuffer() []view.Metric {
	metrics := make([]view.Metric, len(w.buffer))
	iter := 0

	for _, metric := range w.buffer {
		metrics[iter] = metric
		iter++
	}

	w.buffer = make(map[string]view.Metric)

	return metrics
}
