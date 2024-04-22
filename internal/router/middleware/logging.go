package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	dataLength int
}

func (rec *responseRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *responseRecorder) Write(data []byte) (int, error) {
	rec.dataLength = len(data)
	return rec.ResponseWriter.Write(data)
}

func Logger(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// create new instance of responseRecorder
		rec := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusInternalServerError,
			dataLength:     0,
		}
		// start timer
		startTime := time.Now().UnixMilli()
		next.ServeHTTP(rec, r)
		// stop timer after execution
		deltaTime := time.Now().UnixMilli() - startTime
		// print log message
		slog.Info(
			"incoming request",
			slog.String("method", r.Method),
			slog.String("url", r.RequestURI),
			slog.Int64("time_taken_ms", deltaTime),
			slog.Group(
				"response",
				slog.Int("status", rec.statusCode),
				slog.Int("body_length", rec.dataLength),
			),
		)
	})
}
