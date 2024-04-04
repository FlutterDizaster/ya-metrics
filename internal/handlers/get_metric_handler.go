package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type GetMetricHandler struct {
	storage GetMetricStorage
}

func NewGetMetricHandler(storage GetMetricStorage) GetMetricHandler {
	return GetMetricHandler{
		storage: storage,
	}
}

func (h GetMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	kind := chi.URLParam(r, "kind")
	name := chi.URLParam(r, "name")

	value, err := h.storage.GetMetricValue(kind, name)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

	_, err = w.Write([]byte(value))
	if err != nil {
		http.Error(w, "", http.StatusTeapot)
	}
}
