package api

import "net/http"

func (api *API) pingHandler(w http.ResponseWriter, _ *http.Request) {
	err := api.storage.Ping()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
