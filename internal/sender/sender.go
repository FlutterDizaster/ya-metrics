package sender

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/go-resty/resty/v2"
)

type Settings struct {
	Addr           string
	ReportInterval int
	RetryCount     int
	RetryInterval  time.Duration
}

type Sender struct {
	metricsBuffer  []view.Metric
	endpointAddr   string
	reportInterval int
	client         *resty.Client
	cond           *sync.Cond
}

func NewSender(settings *Settings) *Sender {
	slog.Debug("Creating sender")
	sender := &Sender{
		metricsBuffer:  make([]view.Metric, 0),
		endpointAddr:   fmt.Sprintf("http://%s/update", settings.Addr),
		reportInterval: settings.ReportInterval,
		client:         resty.New(),
		cond:           sync.NewCond(&sync.Mutex{}),
	}
	sender.client.SetRetryCount(settings.RetryCount)
	sender.client.SetRetryWaitTime(settings.RetryInterval)
	return sender
}

func (s *Sender) Start(_ context.Context) {
	slog.Debug("Start sending metrics")
	ticker := time.NewTicker(time.Duration(s.reportInterval) * time.Second)
	for {
		s.cond.L.Lock()
		// Ждем добавления метрик
		slog.Debug("Waiting metrics")

		s.cond.Wait()

		go s.sendAll(s.metricsBuffer)

		s.cond.L.Unlock()

		// Ждем тикер
		<-ticker.C
	}
}

func (s *Sender) sendAll(metrics []view.Metric) {
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

func (s *Sender) AddMetrics(metrics []view.Metric) {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()

	// s.metricsBuffer = append(s.metricsBuffer, metrics...)
	s.metricsBuffer = metrics

	s.cond.Broadcast()
}
