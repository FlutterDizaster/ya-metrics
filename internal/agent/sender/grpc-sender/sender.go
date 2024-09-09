package grpcsender

import (
	"context"
	"log/slog"
	"time"

	"github.com/FlutterDizaster/ya-metrics/internal/agent/sender"
	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/FlutterDizaster/ya-metrics/pkg/workerpool"
	pb "github.com/FlutterDizaster/ya-metrics/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Settings struct {
	Addr           string        // Адрес сервера агрегации метрик
	ReportInterval time.Duration // Интервал между отправками метрик
	Buf            sender.Buffer // Буфер метрик
	RateLimit      int           // Максимальное кол-во запросов в секунду
}

type Sender struct {
	endpointAddr   string
	client         pb.MetricsServiceClient
	reportInterval time.Duration
	buf            sender.Buffer
	wpool          workerpool.WorkerPool
}

func New(settings Settings) *Sender {
	return &Sender{
		endpointAddr:   settings.Addr,
		reportInterval: settings.ReportInterval,
		buf:            settings.Buf,
		wpool:          *workerpool.New(settings.RateLimit),
	}
}

func (s *Sender) Start(ctx context.Context) error {
	slog.Debug("Sender", slog.String("status", "start"))
	// Создание подключения
	conn, err := grpc.NewClient(
		s.endpointAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("failed create connection", "error", err)
		return err
	}
	defer conn.Close()

	client := pb.NewMetricsServiceClient(conn)
	s.client = client

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
	metrics, err := s.buf.Pull()
	if err != nil {
		slog.Error("Sender", "error", err)
		return
	}

	// Маршалинг метрик
	pbMetrics := view.MarshalGRPCMetrics(metrics)

	req := &pb.AddMetricsRequest{
		Metrics: pbMetrics,
	}

	// Отправка метрик
	_, err = s.client.AddMetrics(ctx, req)
	if err != nil {
		slog.Error("Sender", "error", err)
	}
}
