package sender

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

type MetricStorage interface {
	GetAll() []struct {
		Name  string
		Kind  string
		Value string
	}
}

type Sender struct {
	serverAddr     string
	serverPort     string
	reportInterval int
	storage        MetricStorage
	client         http.Client
}

func NewSender(port string, addr string, reportInterval int, storage MetricStorage) Sender {
	return Sender{
		serverAddr:     addr,
		serverPort:     port,
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
	metrics := s.storage.GetAll()
	for _, metric := range metrics {
		go s.sendMetric(metric.Name, metric.Kind, metric.Value)
	}
}

func (s *Sender) sendMetric(name string, kind string, value string) {
	//creating url
	url := fmt.Sprintf("http://%s:%s/update/%s/%s/%s", s.serverAddr, s.serverPort, kind, name, value)

	//creating request
	req, err := http.NewRequest(http.MethodPost, url, http.NoBody)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := s.client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("the server returned an error code \"%s\" when trying to send a metric {name: %s, kind %s, value %s}", resp.Status, name, kind, value)
	}
}
