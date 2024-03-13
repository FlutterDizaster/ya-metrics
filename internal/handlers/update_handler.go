package handlers

import (
	"net/http"
	"strings"
)

type MetricStorage interface {
	AddMetricValue(kind string, name string, value string) error
}

type UpdateHandler struct {
	storage MetricStorage
}

func NewUpdateHandler(storage MetricStorage) UpdateHandler {
	return UpdateHandler{
		storage: storage,
	}
}

func (h UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//Check method type
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//Change later
	//Check request format
	urlParts := strings.Split(r.URL.Path, "/")
	if len(urlParts) != 5 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	//Change later
	//Try to add metric to the storage
	if err := h.storage.AddMetricValue(urlParts[2], urlParts[3], urlParts[4]); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Write response
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
