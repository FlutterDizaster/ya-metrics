package interceptors

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type LoggerInterceptor struct {
}

var _ Interceptor = &LoggerInterceptor{}

func (l *LoggerInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		startTime := time.Now()

		// Вызов обработчика запроса
		resp, err := handler(ctx, req)

		// Подсчет времени обработки запроса
		duration := time.Since(startTime)

		addr := "unknown"

		p, ok := peer.FromContext(ctx)
		if !ok {
			slog.Error("unable to get client IP")
		} else {
			addr = p.Addr.String()
		}

		// Запись лога
		slog.Info(
			"incoming unary gRPC request",
			slog.String("method", info.FullMethod),
			slog.String("from", addr),
			slog.Duration("duration", duration),
			slog.Any("error", err),
		)

		return resp, err
	}
}

func (l *LoggerInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, stream)
	}
}
