package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (r *Router) updateHandler(w http.ResponseWriter, req *http.Request) {
	// Try to add metric to the storage
	kind := chi.URLParam(req, "kind")
	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")

	if err := r.storage.AddMetricValue(kind, name, value); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
