package memory

import (
	"bufio"
	"errors"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

var (
	errNotFound  = errors.New("metric not found")
	errWrongType = errors.New("wrong metric type")
)

// Тип Settings используется для хранения настроек хранилища метрик.
type Settings struct {
	StoreInterval   int    // Интервал между записями в файл бекапа
	FileStoragePath string // Путь к файлу бекапа
	Restore         bool   // Флаг восстановления бекапа
}

// Тип MetricStorage используется для хранения метрик в оперативной памяти во время исполнения
// и бекапа метрик по заданным правилам.
type MetricStorage struct {
	storeInterval   int
	fileStoragePath string
	file            *os.File
	writer          *bufio.Writer
	metrics         map[string]view.Metric
	cond            *sync.Cond
	awaiting        atomic.Bool
}

// Функция фабрика для создания нового экземпляра MetricStorage.
// В качестве параметров принимает настройки хранилища.
func New(settings *Settings) (*MetricStorage, error) {
	slog.Debug("Creating DB storage")
	ms := &MetricStorage{
		storeInterval:   settings.StoreInterval,
		fileStoragePath: settings.FileStoragePath,
		metrics:         make(map[string]view.Metric),
		cond:            sync.NewCond(&sync.Mutex{}),
	}

	ms.awaiting.Store(false)

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
		return nil, err
	}

	ms.file = file
	ms.writer = bufio.NewWriter(file)

	slog.Debug("Storage created")
	return ms, nil
}

// Метод добавления метрик в хранилище.
// В качестве параметров принимает метрики.
// Возвращает обновленные метрики.
// В случае ошибки возвращает ошибку.
func (ms *MetricStorage) AddMetrics(metrics ...view.Metric) ([]view.Metric, error) {
	// Блокировка mutex в cond, чтобы избежать чтения данных при бекапе.
	ms.cond.L.Lock()
	defer func() {
		// Оповещение фенкции бекапа о том, что можно продолжать выполнение программы.
		ms.cond.Broadcast()
		ms.cond.L.Unlock()
	}()

	result := make([]view.Metric, 0, len(metrics))

	for i := range metrics {
		metric := metrics[i]

		var err error
		switch metric.MType {
		case view.KindGauge:
			metric, err = ms.addGauge(metric)
		case view.KindCounter:
			metric, err = ms.addCounter(metric)
		default:
			return nil, errWrongType
		}
		if err != nil {
			return nil, err
		}

		result = append(result, metric)
	}
	return result, nil
}

// Метод получения метрики из хранилища.
// Возвращает ошибку в случае если метрика не найдена или у метрики с ID = name другой тип.
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

// Возвращает слайс всех хранящихся метрик.
func (ms *MetricStorage) ReadAllMetrics() ([]view.Metric, error) {
	ms.cond.L.Lock()
	defer ms.cond.L.Unlock()

	return ms.getAllMetrics(), nil
}

// Хелпер фенкция для добавления метрики типа kindCounter.
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

// Хелпер фенкция для добавления метрики типа kindGauge.
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
