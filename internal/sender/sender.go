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

type Sender struct {
	endpointAddr   string
	reportInterval int
	storage        MetricStorage
	client         *resty.Client
}

func NewSender(addr string, reportInterval int, storage MetricStorage) Sender {
	return Sender{
		endpointAddr:   fmt.Sprintf("http://%s/update", addr),
		reportInterval: reportInterval,
		storage:        storage,
		client:         resty.New(),
	}
}

func (s *Sender) Start(ctx context.Context) {
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
	metrics := s.storage.PullAllMetrics()
	for _, metric := range metrics {
		go s.sendMetric(metric)
	}
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
