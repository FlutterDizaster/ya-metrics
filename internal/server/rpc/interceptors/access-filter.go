package interceptors

import (
	"context"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type AccessFilterInterceptor struct {
	// TrustedSubnet - подсеть, в которой разрешен доступ
	TrustedSubnet *net.IPNet
}

var _ Interceptor = &AccessFilterInterceptor{}

func (i *AccessFilterInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Получение информации о клиенте
		p, ok := peer.FromContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "unable to get client IP")
		}

		// Получение IP адреса клиента и удаление порта
		addr := p.Addr.String()
		var err error
		if strings.Contains(addr, ":") {
			addr, _, err = net.SplitHostPort(addr)
			if err != nil {
				return nil, status.Error(codes.Unauthenticated, "unable to get client IP")
			}
		}

		// Проверка подсети
		if i.TrustedSubnet != nil {
			if !i.TrustedSubnet.Contains(net.ParseIP(addr)) {
				return nil, status.Error(codes.PermissionDenied, "access denied")
			}
		}

		return handler(ctx, req)
	}
}

func (i *AccessFilterInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, stream)
	}
}
