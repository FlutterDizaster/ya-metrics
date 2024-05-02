package memory

import (
	"bufio"
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"

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

// // type MSI interface {
// 	AddMetric(view.Metric) error
// 	GetMetric(kind string, name string) (view.Metric, error)
// 	ReadAllMetrics() []view.Metric
// 	PullAllMetrics() []view.Metric
// }

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
}

func NewMetricStorage(settings *Settings) *MetricStorage {
	ms := &MetricStorage{
		storeInterval:   settings.StoreInterval,
		fileStoragePath: settings.FileStoragePath,
		metrics:         make(map[string]view.Metric),
		cond:            sync.NewCond(&sync.Mutex{}),
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

func (ms *MetricStorage) StartBackups(ctx context.Context) {
	slog.Debug("Start backup service")
	defer slog.Debug("Backup service successfully stopped")
	if ms.storeInterval == 0 {
		for {
			select {
			case <-ctx.Done():
				ms.backup(true)
				return
			default:
				ms.backup(false)
			}
		}
	} else {
		ticker := time.NewTicker(time.Duration(ms.storeInterval) * time.Second)
		for {
			select {
			case <-ctx.Done():
				ms.backup(true)
				ticker.Stop()
				return
			case <-ticker.C:
				ms.backup(false)
			}
		}
	}
}

func (ms *MetricStorage) backup(skipWait bool) {
	ms.cond.L.Lock()
	defer ms.cond.L.Unlock()

	// slog.Debug("Waiting metrics for backup")
	if !skipWait {
		slog.Debug("Skip waiting cond")
		// ms.cond.Wait()
	}

	slog.Debug("Creating backup", slog.String("destination", ms.fileStoragePath))

	err := ms.saveToFile()
	if err != nil {
		slog.Error("backup error", "error", err)
	}

	slog.Debug("Backup created")
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

func (ms *MetricStorage) PullAllMetrics() []view.Metric {
	ms.cond.L.Lock()
	defer ms.cond.L.Unlock()

	metrics := ms.getAllMetrics()

	ms.metrics = make(map[string]view.Metric)

	return metrics
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

func (ms *MetricStorage) loadFromFile() error {
	slog.Debug("Loading backup", slog.String("source", ms.fileStoragePath))
	// Открытие файла для чтения
	file, err := os.OpenFile(ms.fileStoragePath, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Создание сканера
	scanner := bufio.NewScanner(file)

	// Проход по всем строкам файла
	for {
		if !scanner.Scan() {
			return scanner.Err()
		}

		// Чтение строки файла
		data := scanner.Bytes()

		// Анмаршалинг строки
		metric := view.Metric{}
		err = metric.UnmarshalJSON(data)
		if err != nil {
			return err
		}

		// Сохранение метрики в буфер
		ms.metrics[metric.ID] = metric
	}
}

func (ms *MetricStorage) saveToFile() error {
	// Очистка файла
	err := ms.file.Truncate(0)
	if err != nil {
		return err
	}
	// Проход по всем метрикам
	for _, metric := range ms.metrics {
		// Маршалинг метрики в JSON
		var bmetric []byte
		bmetric, err = metric.MarshalJSON()
		if err != nil {
			slog.Error("marshaling error", "error", err)
			return err
		}
		slog.Debug("Writing new entry to backup", slog.String("content", string(bmetric)))
		// Запись метрики в файл
		_, err = ms.writer.Write(bmetric)
		if err != nil {
			slog.Error("writing to file error", "error", err)
			return err
		}
		// Добавление переноса строки
		err = ms.writer.WriteByte('\n')
		if err != nil {
			slog.Error("writing to file error", "error", err)
			return err
		}
	}
	return ms.writer.Flush()
}
