package sender

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log/slog"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/go-resty/resty/v2"
)

// TODO: Прокинуть контекст в resty для Graceful Shutdown

type Settings struct {
	Addr          string
	RetryCount    int
	RetryInterval time.Duration
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
		endpointAddr:  fmt.Sprintf("http://%s/update", settings.Addr),
		client:        resty.New(),
	}
	sender.client.SetRetryCount(settings.RetryCount)
	sender.client.SetRetryWaitTime(settings.RetryInterval)
	return sender
}

func (s *Sender) SendMetrics(metrics []view.Metric) {
	slog.Debug("Sending metrics")

	for _, metric := range metrics {
		s.sendMetric(metric)
	}
	slog.Debug("Metrics sended")
}

func (s *Sender) sendMetric(metric view.Metric) {
	// Marshal метрики
	metricBytes, err := metric.MarshalJSON()
	if err != nil {
		slog.Error("marshaling error", "error", err)
		return
	}

	// Создание запроса
	req := s.client.R().
		SetHeader("Content-Type", "application/json")

	// Сжатие метрики
	data, err := compressData(metricBytes)
	if err != nil {
		req.SetBody(metricBytes)
	} else {
		req.SetHeader("Content-Encoding", "gzip").
			SetBody(data)
	}

	// Отправка запроса
	resp, err := req.Post(s.endpointAddr)

	slog.Info(
		"request send",
		"error", err,
		"status", resp.StatusCode(),
		"metric", metric.ID,
		"value", metric.StringValue(),
	)
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
