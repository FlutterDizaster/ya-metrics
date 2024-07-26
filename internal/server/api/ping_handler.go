package api

import "net/http"

// Handler для проверки соединения с базой данных.
//
// Swagger описание:
// @Summary Ping
// @Description Ping DB donnection
// @Tags health
// @Produce text/plain
// @Success 200 {string} string "OK"
// @Failure 500 {string} string "Error"
// @Router /ping [get]
// Конец Swagger описания.
func (api *API) pingHandler(w http.ResponseWriter, _ *http.Request) {
	err := api.storage.Ping()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
