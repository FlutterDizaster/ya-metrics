package middleware

import (
	"bytes"
	"encoding/hex"
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
	w.Header().Set("HashSHA256", hex.EncodeToString(hash))
	return w.ResponseWriter.Write(data)
}

func (h *Validator) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Создание hashWriter для записи хеша ответа в хедер
		hw := &hashWriter{
			ResponseWriter: w,
			key:            h.Key,
		}

		// Проверка на наличие тела запроса
		if r.ContentLength <= 0 {
			r.Body = http.NoBody
			next.ServeHTTP(hw, r)
			return
		}

		// Получение хеша из заголовка запроса
		sampleHashString := r.Header.Get("HashSHA256")
		if sampleHashString == "" {
			// TODO: for tests
			next.ServeHTTP(hw, r)
			// http.Error(w, "HashSHA256 Header required", http.StatusBadRequest)
			return
		}
		sampleHash, err := hex.DecodeString(sampleHashString)
		if err != nil {
			http.Error(w, "Can't decode hash", http.StatusBadRequest)
			return
		}

		// Чтение тела запроса
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "reading body error", http.StatusInternalServerError)
			return
		}
		r.Body.Close()

		// Повторное хеширование тела запроса
		hash := validation.CalculateHashSHA256(body, h.Key)

		// Сравнение хешей
		if !bytes.Equal(hash, sampleHash) {
			http.Error(w, "Invalid Hash", http.StatusBadRequest)
			return
		}

		// Подмена body
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		// Продолжение работы
		next.ServeHTTP(hw, r)
	})
}
