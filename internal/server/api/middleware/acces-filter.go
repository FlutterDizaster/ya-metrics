package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

// AccessFilter является middleware функцией для использования совместно с chi роутером.
// Используется для контроля доступа по указанной подсети.
// Если IP адрес не принадлежит подсети, то возвращается статус 403.
// Если IP адрес принадлежит подсети, то производится обработка запроса.
// По умолчанию проверяет Request.RemoteAddr.
// Проверку IP адреса по заголовкам X-Real-IP и X-Forwarded-For можно выключить с помощью параметра GetIPFromHeaders.
type AccessFilter struct {
	// TrustedSubnet - подсеть, в которой разрешен доступ
	TrustedSubnet *net.IPNet
	// GetIPFromHeaders - проверять ли IP адреса по заголовкам X-Real-IP и X-Forwarded-For
	GetIPFromHeaders bool
}

// Handle - обработка запроса.
func (f *AccessFilter) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteIP, err := f.getRemoteIP(r)
		if err != nil {
			http.Error(w, "Failed to get remote IP", http.StatusBadRequest)
			return
		}
		if !f.TrustedSubnet.Contains(remoteIP) {
			http.Error(w, "access denied", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// getRemoteIP - возвращает IP адрес клиента.
func (f *AccessFilter) getRemoteIP(r *http.Request) (net.IP, error) {
	if f.GetIPFromHeaders {
		return f.getIPFromHeaders(r)
	}
	return f.getIPFromRemoteAddr(r)
}

// getIPFromHeaders - возвращает IP адрес клиента из заголовков.
func (f *AccessFilter) getIPFromHeaders(r *http.Request) (net.IP, error) {
	addr := r.Header.Get("X-Real-IP")

	if addr == "" {
		ips := r.Header.Get("X-Forwarded-For")
		ipStrs := strings.Split(ips, ",")
		addr = ipStrs[0]
	}

	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP: %s", addr)
	}

	return ip, nil
}

// getIPFromRemoteAddr - возвращает IP адрес клиента по RemoteAddr.
func (f *AccessFilter) getIPFromRemoteAddr(r *http.Request) (net.IP, error) {
	addr := r.RemoteAddr

	ipStr, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, err
	}

	return ip, nil
}
