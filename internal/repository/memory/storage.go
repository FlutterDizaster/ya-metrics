package memory

import (
	"bufio"
	"errors"
	"log/slog"
	"os"
	"sync"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

var (
	errNotFound  = errors.New("metric not found")
	errWrongType = errors.New("wrong metric type")
)

const (
	kindGauge   = "gauge"
	kindCounter = "counter"
)

type Settings struct {
	StoreInterval   int
	FileStoragePath string
	Restore         bool
}

type MetricStorage struct {
	storeInterval   int
	fileStoragePath string
	file            *os.File
	writer          *bufio.Writer
	metrics         map[string]view.Metric
	cond            *sync.Cond
	awmtx           sync.Mutex
	awaiting        bool
}

func NewMetricStorage(settings *Settings) *MetricStorage {
	ms := &MetricStorage{
		storeInterval:   settings.StoreInterval,
		fileStoragePath: settings.FileStoragePath,
		metrics:         make(map[string]view.Metric),
		cond:            sync.NewCond(&sync.Mutex{}),
		awaiting:        false,
	}

	if settings.Restore {
		err := ms.loadFromFile()
		if err != nil {
			slog.Error("error reading backup file", "error", err)
			slog.Info("Skipping loading backup...")
		}
	}

	file, err := os.OpenFile(ms.fileStoragePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		slog.Error("error opening file", "error", err)
		os.Exit(1)
	}

	ms.file = file
	ms.writer = bufio.NewWriter(file)

	return ms
}

func (ms *MetricStorage) AddMetric(metric view.Metric) (view.Metric, error) {
	ms.cond.L.Lock()
	defer func() {
		ms.cond.Broadcast()
		ms.cond.L.Unlock()
	}()

	switch metric.MType {
	case kindGauge:
		return ms.addGauge(metric)
	case kindCounter:
		return ms.addCounter(metric)
	default:
		return metric, errWrongType
	}
}

func (ms *MetricStorage) GetMetric(kind string, name string) (view.Metric, error) {
	ms.cond.L.Lock()
	defer ms.cond.L.Unlock()

	metric, ok := ms.metrics[name]
	if !ok {
		return view.Metric{}, errNotFound
	}

	if metric.MType != kind {
		return view.Metric{}, errWrongType
	}

	return metric, nil
}

func (ms *MetricStorage) ReadAllMetrics() []view.Metric {
	ms.cond.L.Lock()
	defer ms.cond.L.Unlock()

	return ms.getAllMetrics()
}

func (ms *MetricStorage) addCounter(metric view.Metric) (view.Metric, error) {
	oldMetric, ok := ms.metrics[metric.ID]
	if !ok {
		ms.metrics[metric.ID] = metric
		return metric, nil
	}

	if oldMetric.MType != metric.MType {
		return metric, errWrongType
	}

	delta := *oldMetric.Delta + *metric.Delta
	metric.Delta = &delta

	ms.metrics[metric.ID] = metric

	return metric, nil
}

func (ms *MetricStorage) addGauge(metric view.Metric) (view.Metric, error) {
	oldMetric, ok := ms.metrics[metric.ID]
	if !ok {
		ms.metrics[metric.ID] = metric
		return metric, nil
	}

	if oldMetric.MType != metric.MType {
		return metric, errWrongType
	}

	ms.metrics[metric.ID] = metric

	return metric, nil
}

func (ms *MetricStorage) getAllMetrics() []view.Metric {
	metrics := make([]view.Metric, len(ms.metrics))
	iter := 0

	for _, metric := range ms.metrics {
		metrics[iter] = metric
		iter++
	}

	return metrics
}
