package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// responseRecorder выступает оберткой над http.ResponseWriter.
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	wroteBody bool
}

// Write переопределение функции http.ResponseWriter.Write([]byte).
func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", http.DetectContentType(data))
	}
	w.wroteBody = true

	w.Header().Set("Content-Encoding", "gzip")

	return w.Writer.Write(data)
}

// GzipCompressor является middleware функцией для использования совместно с chi роутером.
// Сжимает тело ответа, если клиент принимает его в таком виде.
func GzipCompressor(next http.Handler) http.Handler {
	pool := gzipCompressorPool()

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(rw, r)
		}

		i := pool.Get()
		w, ok := i.(*gzip.Writer)
		if !ok {
			http.Error(
				rw,
				fmt.Sprintf("error getting gzip writer from pool: %s", i.(error)),
				http.StatusInternalServerError,
			)
			return
		}
		w.Reset(rw)

		grw := &gzipResponseWriter{
			Writer:         w,
			ResponseWriter: rw,
		}

		defer func() {
			if !grw.wroteBody {
				if grw.Header().Get("Content-Encoding") == "gzip" {
					grw.Header().Del("Content-Encoding")
				}
			}
			w.Close()
			pool.Put(w)
		}()

		next.ServeHTTP(grw, r)
	})
}

func gzipCompressorPool() sync.Pool {
	return sync.Pool{
		New: func() any {
			w, err := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
			if err != nil {
				return err
			}
			return w
		},
	}
}
