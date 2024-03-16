package handlers

import "net/http"

type GetMetricHandler struct {
	storage GetMetricStorage
}

func NewGetMetricHandler(storage GetMetricStorage) GetMetricHandler {
	return GetMetricHandler{
		storage: storage,
	}
}

func (h GetMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
