package worker

import (
	"context"
	"errors"
	"sync"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

var (
	errNilPointer = errors.New("nil pointer error")
)

type Collector interface {
	CollectMetrics() []view.Metric
}

type Sender interface {
	SendMetrics([]view.Metric)
}

type Settings struct {
	Collector      Collector
	Sender         Sender
	ReportInterval int
	PollInterval   int
}

type Worker struct {
	collector      Collector
	sender         Sender
	buffer         map[string]view.Metric
	reportInterval int
	pollInterval   int
	cond           *sync.Cond
	wg             sync.WaitGroup
}

func NewWorker(settings *Settings) (*Worker, error) {
	if settings.Sender == nil || settings.Collector == nil {
		return &Worker{}, errNilPointer
	}

	worker := &Worker{
		collector:      settings.Collector,
		sender:         settings.Sender,
		reportInterval: settings.ReportInterval,
		pollInterval:   settings.PollInterval,
		buffer:         make(map[string]view.Metric),
		cond:           sync.NewCond(&sync.Mutex{}),
	}

	return worker, nil
}

func (w *Worker) Start(ctx context.Context) {
	w.wg.Add(1)
	go func() {
		w.startCollecting(ctx)
		w.wg.Done()
	}()

	w.wg.Add(1)
	go func() {
		w.startSending(ctx)
		w.wg.Done()
	}()

	w.wg.Wait()
}
