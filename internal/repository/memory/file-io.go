package memory

import (
	"bufio"
	"io"
	"log/slog"
	"os"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

// Метод загружающий метрики в хранилище из файла.
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

// Метод записывающий метрики в файл.
func (ms *MetricStorage) saveToFile() error {
	// Очистка файла
	err := ms.file.Truncate(0)
	if err != nil {
		return err
	}
	_, err = ms.file.Seek(0, io.SeekStart)
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
