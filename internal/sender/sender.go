package sender

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

type MetricStorage interface {
	PullAllMetrics() []struct {
		Name  string
		Kind  string
		Value string
	}
}

type Sender struct {
	serverAddr     string
	reportInterval int
	storage        MetricStorage
	client         http.Client
}

func NewSender(addr string, reportInterval int, storage MetricStorage) Sender {
	return Sender{
		serverAddr:     addr,
		reportInterval: reportInterval,
		storage:        storage,
		client:         http.Client{},
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
		go s.sendMetric(metric.Name, metric.Kind, metric.Value)
	}
}

func (s *Sender) sendMetric(name string, kind string, value string) {
	// creating url
	url := fmt.Sprintf("http://%s/update/%s/%s/%s", s.serverAddr, kind, name, value)

	// creating request
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, url, http.NoBody)
	if err != nil {
		log.Printf("unexpected error in sendMetric function\n%s", err)
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("connection error \"%s\" when trying to send a metric %s", err, name)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		//TODO: add error processing
		log.Printf("unexpected status code \"%s\" when trying to send a metric name: %s", resp.Status, name)
	}
}
