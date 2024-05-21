package buffer

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

var (
	errBufferClosed = errors.New("Buffer closed")
)

type Buffer struct {
	metrics map[string]view.Metric
	cond    sync.Cond
	ready   atomic.Bool
	closed  atomic.Bool
}

func New() *Buffer {
	return &Buffer{
		metrics: make(map[string]view.Metric),
		cond:    *sync.NewCond(&sync.Mutex{}),
	}
}

func (b *Buffer) Close() {
	b.closed.Store(true)
}

func (b *Buffer) Put(metrics []view.Metric) error {
	if b.closed.Load() {
		return errBufferClosed
	}

	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	for i := range metrics {
		id := metrics[i].ID
		switch metrics[i].MType {
		case view.KindCounter:
			old, ok := b.metrics[id]
			if !ok {
				b.metrics[id] = metrics[i]
				continue
			}
			delta := *metrics[i].Delta + *old.Delta
			metrics[i].Delta = &delta
			b.metrics[id] = metrics[i]
		case view.KindGauge:
			b.metrics[id] = metrics[i]
		}
	}

	b.ready.Store(true)
	b.cond.Broadcast()
	return nil
}

func (b *Buffer) Pull() ([]view.Metric, error) {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	// Ожидание метрик
	for !b.ready.Load() {
		// Если буфер закрыт, выход из функции
		if b.closed.Load() {
			return []view.Metric{}, errBufferClosed
		}
		b.cond.Wait()
	}

	metrics := make([]view.Metric, 0, len(b.metrics))
	for k := range b.metrics {
		metrics = append(metrics, b.metrics[k])
	}

	b.metrics = make(map[string]view.Metric)

	return metrics, nil
}
