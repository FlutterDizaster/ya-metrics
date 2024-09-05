package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"log/slog"
	"net/http"
)

// RSADecoder является middleware функцией для использования совместно с chi роутером.
// Расшифровывает тело запроса, если клиент отправил его в таком виде.
type RSADecoder struct {
	Key *rsa.PrivateKey
}

func (d *RSADecoder) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			next.ServeHTTP(w, r)
			return
		}

		// чтение тела запроса
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Декодирование тела запроса
		body, err = rsa.DecryptPKCS1v15(nil, d.Key, body)
		if err != nil {
			slog.Error("rsa decoder error", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Подмена body
		r.Body = io.NopCloser(bytes.NewReader(body))
		next.ServeHTTP(w, r)
	})
}
