package router

import (
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
	}

	// добавление метрики в репозиторий
	if nErr := r.storage.AddMetric(*metric); nErr != nil {
		http.Error(w, nErr.Error(), http.StatusBadRequest)
		return
	}

	// записываем ответ
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
