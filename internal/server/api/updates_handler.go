package api

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

func (api *API) updateBatchHandler(w http.ResponseWriter, r *http.Request) {
	var metrics view.Metrics
	var buf bytes.Buffer

	// Чтение тела запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		slog.Error("Reading error", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Unmarshal
	if err = metrics.UnmarshalJSON(buf.Bytes()); err != nil {
		slog.Error("UnmarshalJSON error", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Добавление метрики в репозиторий
	if metrics, err = api.storage.AddMetrics(metrics...); err != nil {
		slog.Error("AddBatchMetrics error", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Marshal ответа
	resp, err := metrics.MarshalJSON()
	if err != nil {
		slog.Error(
			"marshaling error",
			"message", err,
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// записываем ответ
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(resp); err != nil {
		slog.Error("writing response error", "message", err)
		http.Error(w, fmt.Sprintf("write metric error: %s", err), http.StatusInternalServerError)
		return
	}
}
