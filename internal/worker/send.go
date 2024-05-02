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
	// Создание контекста сендера
	// TODO: Переделать контекст сендера, а то как-то плохо оно работает
	senderCtx, senderStopCtx := context.WithCancel(context.TODO())
	// Первая отправка метрик
	w.send(senderCtx)
	for {
		select {
		case <-ctx.Done():
			slog.Debug("Stopping sender...")
			w.send(senderCtx)
			time.AfterFunc(gracefullPeriodSec*time.Second, senderStopCtx)
			ticker.Stop()
			// Выходим из функции
			return
		case <-ticker.C:
			w.send(senderCtx)
		}
	}
}

func (w *Worker) send(ctx context.Context) {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	// Ждем добавления метрик
	slog.Debug("Waiting metrics")

	w.cond.Wait()

	w.wg.Add(1)
	go func() {
		w.sender.SendMetrics(ctx, w.pullBuffer())
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
