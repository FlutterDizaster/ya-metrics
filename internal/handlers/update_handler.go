package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type UpdateHandler struct {
	storage AddMetricStorage
}

func NewUpdateHandler(storage AddMetricStorage) UpdateHandler {
	return UpdateHandler{
		storage: storage,
	}
}

func (h UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Try to add metric to the storage
	kind := chi.URLParam(r, "kind")
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	if err := h.storage.AddMetricValue(kind, name, value); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
