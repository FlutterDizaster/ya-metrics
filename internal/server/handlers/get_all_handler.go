package handlers

import "net/http"

type GetAllHandler struct {
	storage GetAllMetricsStorage
}

func NewGetAllHandler(storage GetAllMetricsStorage) GetAllHandler {
	return GetAllHandler{
		storage: storage,
	}
}

func (h GetAllHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
