package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handler для обноваления состояния метрики в репозитории.
func (r *Router) updateHandler(w http.ResponseWriter, req *http.Request) {
	// парсинг URL запроса
	kind := chi.URLParam(req, "kind")
	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")

	// добавление метрики в репозиторий
	if err := r.storage.AddMetricValue(kind, name, value); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// записываем ответ
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
