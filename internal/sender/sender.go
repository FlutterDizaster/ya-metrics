package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/go-resty/resty/v2"
)

type MetricStorage interface {
	PullAllMetrics() []view.Metric
}

type Settings struct {
	Addr           string
	ReportInterval int
	Storage        MetricStorage
	RetryCount     int
	RetryInterval  time.Duration
}

type Sender struct {
	endpointAddr   string
	reportInterval int
	storage        MetricStorage
	client         *resty.Client
}

func NewSender(settings *Settings) *Sender {
	slog.Debug("Creating sender")
	sender := &Sender{
		endpointAddr:   fmt.Sprintf("http://%s/update", settings.Addr),
		reportInterval: settings.ReportInterval,
		storage:        settings.Storage,
		client:         resty.New(),
	}
	sender.client.SetRetryCount(settings.RetryCount)
	sender.client.SetRetryWaitTime(settings.RetryInterval)
	return sender
}

func (s *Sender) Start(ctx context.Context) {
	slog.Debug("Start sending metrics")
	for {
		select {
		case <-ctx.Done():
			time.Sleep(1 * time.Second)
			s.sendAll()
			return
		default:
			s.sendAll()
			time.Sleep(time.Duration(s.reportInterval) * time.Second)
		}
	}
}

func (s *Sender) sendAll() {
	slog.Debug("Sending metrics")
	timer := time.Now()

	metrics := s.storage.PullAllMetrics()
	for _, metric := range metrics {
		go s.sendMetric(metric)
	}
	slog.Debug("Metrics sended", "delta time ms", time.Since(timer).Milliseconds())
}

func (s *Sender) sendMetric(metric view.Metric) {
	// Marshal метрики
	metricBytes, err := json.Marshal(metric)
	if err != nil {
		slog.Error("marshaling error", "message", err)
	}

	// Создание запроса
	resp, err := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(metricBytes).
		Post(s.endpointAddr)

	slog.Info(
		"request send",
		"error", err,
		"status", resp.StatusCode(),
		"metric", metric.ID,
		"value", metric.StringValue(),
	)
}
