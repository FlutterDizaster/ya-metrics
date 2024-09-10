package rpc

import (
	"context"
	"log/slog"
	"net"

	"github.com/FlutterDizaster/ya-metrics/internal/server/rpc/interceptors"
	"github.com/FlutterDizaster/ya-metrics/internal/view"
	pb "github.com/FlutterDizaster/ya-metrics/proto"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricsStorage interface {
	AddMetrics(metrics ...view.Metric) ([]view.Metric, error)
}

type Settings struct {
	Storage      MetricsStorage
	Addr         string
	Interceptors []interceptors.Interceptor
}

// MetricsService - gRPC сервис для работы с метриками.
// Используется для коммуникации с клиентом сборщиком метрик.
type MetricsService struct {
	pb.UnimplementedMetricsServiceServer
	storage      MetricsStorage
	addr         string
	interceptors []interceptors.Interceptor
}

// New - создание экземпляра MetricsService.
// В качестве параметров принимает настройки gRPC сервера.
// Возвращает экземпляр MetricsService.
func New(settings Settings) *MetricsService {
	return &MetricsService{
		storage:      settings.Storage,
		addr:         settings.Addr,
		interceptors: settings.Interceptors,
	}
}

// Start - запуск gRPC сервера.
// Блокирует потов выполнения до завершения работы сервиса.
// Завершает работу сервиса при завершении контекста.
// Перед вызовом функции необходимо создать экземпляр MetricsService с помощью New().
func (s *MetricsService) Start(ctx context.Context) error {
	slog.Info("Starting RPC server")

	listen, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	interceptors := make([]grpc.UnaryServerInterceptor, 0, len(s.interceptors))
	for i := range s.interceptors {
		interceptors = append(interceptors, s.interceptors[i].Unary())
	}

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))

	pb.RegisterMetricsServiceServer(srv, s)

	eg := errgroup.Group{}

	eg.Go(func() error {
		return srv.Serve(listen)
	})

	<-ctx.Done()

	srv.GracefulStop()

	return eg.Wait()
}

// AddMetrics - gRPC обработчик добавления метрик в хранилище.
// Метод принимает слайс метрик для послежующего добавления их в репозиторий и возвращает слайс обновленных метрик.
func (s *MetricsService) AddMetrics(
	_ context.Context,
	req *pb.AddMetricsRequest,
) (*pb.AddMetricsResponse, error) {
	metrics := view.UnmarshalGRPCMetrics(req.GetMetrics())

	resutl, err := s.storage.AddMetrics(metrics...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add metrics: %v", err)
	}

	resp := &pb.AddMetricsResponse{
		Metrics: view.MarshalGRPCMetrics(resutl),
	}

	return resp, nil
}
