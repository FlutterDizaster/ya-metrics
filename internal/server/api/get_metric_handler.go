package api

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/go-chi/chi/v5"
)

// Handler для получения значения конкретной метрики по её типу и имени.
func (r *Router) getMetricHandler(w http.ResponseWriter, req *http.Request) {
	// парсинг url запроса для получения типа и имени искомой метрики
	kind := chi.URLParam(req, "kind")
	name := chi.URLParam(req, "name")

	// получение метрики из репозитория
	metric, err := r.storage.GetMetric(kind, name)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	// отдача метрики клиенту
	_, err = w.Write([]byte(metric.StringValue()))
	if err != nil {
		slog.Error("writing response error", "message", err)
		http.Error(w, fmt.Sprintf("write metric error: %s", err), http.StatusInternalServerError)
		return
	}
}

// Handler для получения значения конкретной метрики по её типу и имени в формате JSON.
func (r *Router) getJSONMetricHandler(w http.ResponseWriter, req *http.Request) {
	var reqMetric view.Metric
	var buf bytes.Buffer

	// читаем тело запроса
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Unmarshal тела запроса
	if err = reqMetric.UnmarshalJSON(buf.Bytes()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// получение метрики из репозитория
	metric, err := r.storage.GetMetric(reqMetric.MType, reqMetric.ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	// Marshal ответа
	resp, err := metric.MarshalJSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// записываем ответ
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(resp); err != nil {
		slog.Error("writing response error", "message", err)
		http.Error(w, fmt.Sprintf("write metric error: %s", err), http.StatusInternalServerError)
		return
	}
}
