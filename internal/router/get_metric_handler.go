package router

import (
	"fmt"
	"net/http"

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
		http.Error(w, fmt.Sprintf("write metric error: %s", err), http.StatusInternalServerError)
	}
}
