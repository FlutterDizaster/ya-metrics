package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/FlutterDizaster/ya-metrics/pkg/validation"
	"github.com/FlutterDizaster/ya-metrics/pkg/workerpool"
	"github.com/go-resty/resty/v2"
)

// TODO: Прокинуть контекст в resty для Graceful Shutdown

// Интерфейс для буфера метрик.
type Buffer interface {
	// Метод для вытягивания всех метрик из буфера.
	// Подразумевается, что после вызова буфер будет очищен.
	Pull() ([]view.Metric, error)
}

// Настройки сервиса отправки метрик.
type Settings struct {
	Addr             string        // Адрес сервера агрегации метрик
	RetryCount       int           // Количество повторных попыток отправки метрик
	RetryInterval    time.Duration // Интервал между повторными попытками
	RetryMaxWaitTime time.Duration // Максимальное время ожидания между повторными попытками
	ReportInterval   time.Duration // Интервал между отправками метрик
	Key              string        // Хеш ключ
	Buf              Buffer        // Буфер метрик
	RateLimit        int           // Максимальное кол-во запросов в секунду
}

// Sender - сервис отправки метрик.
// Должен быть создан через New.
type Sender struct {
	endpointAddr   string
	client         *resty.Client
	reportInterval time.Duration
	key            string
	buf            Buffer
	wpool          workerpool.WorkerPool
}

// Фабрика создания экземпляра Sender.
func New(settings Settings) *Sender {
	slog.Debug("Creating sender")
	sender := &Sender{
		endpointAddr:   fmt.Sprintf("http://%s/updates/", settings.Addr),
		client:         resty.New(),
		reportInterval: settings.ReportInterval,
		key:            settings.Key,
		buf:            settings.Buf,
		wpool:          *workerpool.New(settings.RateLimit),
	}
	sender.client.SetRetryCount(settings.RetryCount)
	sender.client.SetRetryWaitTime(settings.RetryInterval)
	sender.client.SetRetryMaxWaitTime(settings.RetryMaxWaitTime)
	return sender
}

// Start - запуск сервиса отправки метрик.
// Блокирует потов выполнения до завершения работы сервиса.
// Завершает работу сервиса при завершении контекста.
func (s *Sender) Start(ctx context.Context) error {
	slog.Debug("Sender", slog.String("status", "start"))
	ticker := time.NewTicker(s.reportInterval)
	slog.Info("Sender started", "report interval", s.reportInterval)

	// Первая отправка метрик
	s.send(ctx)

	// Старт основного цикла
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			lastCtx, lastCancleCtx := context.WithTimeout(context.Background(), 3*time.Second)
			defer lastCancleCtx()
			s.send(lastCtx)
			s.wpool.Close()
			slog.Debug("Sender", slog.String("status", "stop"))
			return nil
		case <-ticker.C:
			s.send(ctx)
		}
	}
}

func (s *Sender) send(ctx context.Context) {
	slog.Debug("Sender", slog.String("status", "sending..."))
	// ПОлучение метрик из буфера агента
	// var metrics view.Metrics
	metrics, err := s.buf.Pull()
	if err != nil {
		slog.Error("Sender", "error", err)
		return
	}

	// Маршалинг метрик
	metricsBytes, err := view.Metrics(metrics).MarshalJSON()
	if err != nil {
		slog.Error("marshaling error", "error", err)
		return
	}

	// Формирование запроса
	req := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetContext(ctx)

	// Подсчет хеша при необходимости
	if s.key != "" {
		slog.Debug("Calculating hash")
		hash := validation.CalculateHashSHA256(metricsBytes, []byte(s.key))
		req.SetHeader("HashSHA256", hex.EncodeToString(hash))
	}

	// Сжатие метрики
	data, err := compressData(metricsBytes)
	if err != nil {
		req.SetBody(metricsBytes)
	} else {
		req.SetHeader("Content-Encoding", "gzip").
			SetBody(data)
	}

	// Отправка запроса
	err = s.wpool.Do(func() {
		resp, errr := req.Post(s.endpointAddr)
		if errr != nil {
			slog.Info("Sender", "error", errr)
		} else {
			slog.Info(
				"Sender",
				slog.String("status", "sended"),
				slog.Int("response_code", resp.StatusCode()),
			)
		}
	})
	if err != nil {
		slog.Error("unexpected sender error", "error", err)
	}
}

func compressData(data []byte) ([]byte, error) {
	buf := &bytes.Buffer{}
	gz, err := gzip.NewWriterLevel(buf, gzip.BestSpeed)
	if err != nil {
		slog.Error("failed init gzip writer", "error", err)
		return []byte{}, err
	}

	_, err = gz.Write(data)
	if err != nil {
		slog.Error("compress error", "error", err)
		return []byte{}, err
	}

	err = gz.Close()
	if err != nil {
		slog.Error("compress error", "error", err)
		return []byte{}, err
	}

	return buf.Bytes(), nil
}
