package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseData struct {
	statusCode int
	dataSize   int
}

// responseRecorder выступает оберткой над http.ResponseWriter.
// Сохраняет status code ответа и кол-во байт тела ответа.
type responseRecorder struct {
	http.ResponseWriter
	responseData *responseData
}

// WriteHeader переопределение функции http.ResponseWriter.WriteHeader(int).
// Сохраняет статус код ответа, затем передает управление функции http.ResponseWriter.WriteHeader(int).
func (r *responseRecorder) WriteHeader(code int) {
	r.ResponseWriter.WriteHeader(code)
	r.responseData.statusCode = code
}

// Write переопределение функции http.ResponseWriter.Write([]byte).
// Сохраняет кол-во байт тела ответа, затем передает управление функции http.ResponseWriter.Write([]byte).
func (r *responseRecorder) Write(data []byte) (int, error) {
	if r.responseData.statusCode == 0 {
		r.WriteHeader(http.StatusOK)
	}
	size, err := r.ResponseWriter.Write(data)

	r.responseData.dataSize += size
	return size, err
}

// Logger является middleware функцией для использования совместно с chi роутером.
// Выводит с помощью slog сообщение с указанием метода запрос, URL адреса, времени выполнения в ms,
// статус код ответа и кол-во байт тела ответа.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// start timer
		startTime := time.Now()

		// create new instance of responseRecorder
		resData := &responseData{
			statusCode: 0,
			dataSize:   0,
		}

		rec := &responseRecorder{
			ResponseWriter: w,
			responseData:   resData,
		}
		next.ServeHTTP(rec, r)
		// stop timer after execution
		deltaTime := time.Since(startTime)
		// print log message
		slog.Info(
			"incoming request",
			slog.String("method", r.Method),
			slog.String("url", r.RequestURI),
			slog.String("accept-encoding", r.Header.Get("Accept-Encoding")),
			slog.Int64("time_taken_ms", deltaTime.Milliseconds()),
			slog.Group(
				"response",
				slog.Int("status", rec.responseData.statusCode),
				slog.Int("body_length", rec.responseData.dataSize),
			),
		)
	})
}
