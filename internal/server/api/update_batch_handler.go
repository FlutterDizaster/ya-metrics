package api

import (
	"bytes"
	"net/http"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

func (api *API) updateBatchHandler(w http.ResponseWriter, r *http.Request) {
	var metrics view.Metrics
	var buf bytes.Buffer

	// Чтение тела запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Unmarshal
	if err = metrics.UnmarshalJSON(buf.Bytes()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Добавление метрики в репозиторий
	if err = api.storage.AddBatchMetrics(metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
