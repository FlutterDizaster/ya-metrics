package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/go-chi/chi/v5"
)

// Handler для обноваления состояния метрики в репозитории.
func (r *Router) updateHandler(w http.ResponseWriter, req *http.Request) {
	// парсинг URL запроса
	kind := chi.URLParam(req, "kind")
	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")

	// создание новой метрики
	metric, err := view.NewMetric(kind, name, value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// добавление метрики в репозиторий
	if _, err = r.storage.AddMetric(*metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// записываем ответ
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (r *Router) updateJSONHandler(w http.ResponseWriter, req *http.Request) {
	var metric view.Metric
	var buf bytes.Buffer

	// читаем тело запроса
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Unmarshal тела запроса
	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// добавление метрики в репозиторий
	if metric, err = r.storage.AddMetric(metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Marshal ответа
	resp, err := json.Marshal(metric)
	if err != nil {
		slog.Error(
			"marshaling error",
			"message", err,
			"metric", metric.ID,
			"value", metric.StringValue(),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// записываем ответ
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Content-Type", "charset=utf-8")
	if _, err = w.Write(resp); err != nil {
		slog.Error("writing response error", "message", err)
		http.Error(w, fmt.Sprintf("write metric error: %s", err), http.StatusInternalServerError)
		return
	}
}
