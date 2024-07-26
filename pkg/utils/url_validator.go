package utils

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

var (
	ErrURLWithoutHost = errors.New("url must contain host")
	ErrNotEmptyScheme = errors.New("url must be without scheme")
	ErrInvalidPort    = errors.New("url contains invalid port")
)

// Функция проверки валидности URL.
// Если URL содержит схему, то возвращает ошибку.
// Если URL не содержит хост, то возвращает ошибку.
// Если URL содержит некорректный порт, то возвращает ошибку.
// Если URL содержит некорректный адрес, то возвращает ошибку.
// Если URL содержит корректный адрес, то возвращает nil.
func ValidateURL(u string) error {
	// is scheme defined?
	if strings.Contains(u, "://") {
		return ErrNotEmptyScheme
	}

	parts := strings.Split(u, ":")
	addr := parts[0]

	// is host empty?
	if len(addr) == 0 {
		return ErrURLWithoutHost
	}

	// is the port valid?
	if len(parts) > 1 {
		if _, err := net.LookupPort("tcp", parts[1]); err != nil {
			return ErrInvalidPort
		}
	}

	// is the endpoint an IP?
	if net.ParseIP(u) != nil {
		return nil
	}

	// is the endpoint localhost?
	if addr == "localhost" {
		return nil
	}

	// is the endpoint valid url?
	if _, err := url.Parse(addr); err == nil {
		return nil
	}

	return nil
}
