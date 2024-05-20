package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/FlutterDizaster/ya-metrics/pkg/validation"
)

type Validator struct {
	Key []byte
}

type hashWriter struct {
	http.ResponseWriter
	key []byte
}

func (w *hashWriter) Write(data []byte) (int, error) {
	// Подсчет хеша
	hash := validation.CalculateHashSHA256(data, w.key)
	// Установка хедера
	w.Header().Set("HashSHA256", string(hash))
	return w.ResponseWriter.Write(data)
}

func (h *Validator) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Создание hashWriter для записи хеша ответа в хедер
		hw := &hashWriter{
			ResponseWriter: w,
			key:            h.Key,
		}

		// Проверка есть ли у запроса тело
		if r.ContentLength <= 0 {
			r.Body = io.NopCloser(nil)
			next.ServeHTTP(hw, r)
			return
		}

		// Получение хеша из заголовка запроса
		var sample []byte
		if h := r.Header.Get("HashSHA256"); h != "" {
			sample = []byte(h)
		} else {
			http.Error(w, "HashSHA256 Header required", http.StatusBadRequest)
			return
		}

		// Чтение тела запроса
		body, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			http.Error(w, "reading body error", http.StatusInternalServerError)
			return
		}

		// Повторное хеширование тела запроса
		hash := validation.CalculateHashSHA256(body, h.Key)

		// Сравнение хешей
		if !bytes.Equal(hash, sample) {
			http.Error(w, "Invalid Hash", http.StatusBadRequest)
			return
		}

		// Подмена body
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		// Продолжение работы
		next.ServeHTTP(hw, r)
	})
}
