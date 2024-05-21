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
	"github.com/go-resty/resty/v2"
)

// TODO: Прокинуть контекст в resty для Graceful Shutdown

type Buffer interface {
	Pull() ([]view.Metric, error)
}

type Settings struct {
	Addr             string
	RetryCount       int
	RetryInterval    time.Duration
	RetryMaxWaitTime time.Duration
	ReportInterval   time.Duration
	Key              string
	Buf              Buffer
}

type Sender struct {
	endpointAddr   string
	client         *resty.Client
	reportInterval time.Duration
	key            string
	buf            Buffer
}

func New(settings Settings) *Sender {
	slog.Debug("Creating sender")
	sender := &Sender{
		endpointAddr:   fmt.Sprintf("http://%s/updates/", settings.Addr),
		client:         resty.New(),
		reportInterval: settings.ReportInterval,
		key:            settings.Key,
		buf:            settings.Buf,
	}
	sender.client.SetRetryCount(settings.RetryCount)
	sender.client.SetRetryWaitTime(settings.RetryInterval)
	sender.client.SetRetryMaxWaitTime(settings.RetryMaxWaitTime)
	return sender
}

func (s *Sender) Start(ctx context.Context) error {
	slog.Debug("Sender", slog.String("status", "start"))
	ticker := time.NewTicker(s.reportInterval)

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
	resp, err := req.Post(s.endpointAddr)
	if err != nil {
		slog.Info("Sender", "error", err)
	} else {
		slog.Info(
			"Sender",
			slog.String("status", "sended"),
			slog.Int("response_code", resp.StatusCode()),
		)
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
