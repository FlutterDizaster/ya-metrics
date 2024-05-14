package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/go-resty/resty/v2"
)

// TODO: Прокинуть контекст в resty для Graceful Shutdown

type Settings struct {
	Addr             string
	RetryCount       int
	RetryInterval    time.Duration
	RetryMaxWaitTime time.Duration
}

type Sender struct {
	metricsBuffer []view.Metric
	endpointAddr  string
	client        *resty.Client
}

func NewSender(settings *Settings) *Sender {
	slog.Debug("Creating sender")
	sender := &Sender{
		metricsBuffer: make([]view.Metric, 0),
		endpointAddr:  fmt.Sprintf("http://%s/updates/", settings.Addr),
		client:        resty.New(),
	}
	sender.client.SetRetryCount(settings.RetryCount)
	sender.client.SetRetryWaitTime(settings.RetryInterval)
	sender.client.SetRetryMaxWaitTime(settings.RetryMaxWaitTime)
	return sender
}

func (s *Sender) SendMetrics(ctx context.Context, metrics []view.Metric) {
	slog.Debug("Sending metrics")

	s.sendBatch(ctx, metrics)
	// for _, metric := range metrics {
	// 	s.sendMetric(ctx, metric)
	// }
	slog.Debug("Metrics sended")
}

func (s *Sender) sendBatch(ctx context.Context, metrics view.Metrics) {
	metricsBytes, err := metrics.MarshalJSON()
	if err != nil {
		slog.Error("marshaling error", "error", err)
		return
	}

	req := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetContext(ctx)

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

	slog.Info(
		"request send",
		"error", err,
		"metrics count", len(metrics),
		"status", resp.StatusCode(),
		// "response", string(resp.Body()),
	)
}

// func (s *Sender) sendMetric(ctx context.Context, metric view.Metric) {
// 	// Marshal метрики
// 	metricBytes, err := metric.MarshalJSON()
// 	if err != nil {
// 		slog.Error("marshaling error", "error", err)
// 		return
// 	}

// 	// Создание запроса
// 	req := s.client.R().
// 		SetHeader("Content-Type", "application/json").
// 		SetContext(ctx)

// 	// Сжатие метрики
// 	data, err := compressData(metricBytes)
// 	if err != nil {
// 		req.SetBody(metricBytes)
// 	} else {
// 		req.SetHeader("Content-Encoding", "gzip").
// 			SetBody(data)
// 	}

// 	// Отправка запроса
// 	resp, err := req.Post(s.endpointAddr)

// 	slog.Info(
// 		"request send",
// 		"error", err,
// 		"status", resp.StatusCode(),
// 		"metric", metric.ID,
// 		"value", metric.StringValue(),
// 	)
// }

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
