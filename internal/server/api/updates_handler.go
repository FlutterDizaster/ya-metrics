package api

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

// updateBatchHandler обрабатывает POST-запросы на добавление множества метрик в репозиторий.
//
// Swagger описание:
// @Summary Update metrics
// @Description Update metrics in DB
// @Tags metrics
// @Produce json
// @Param metrics body []view.Metric true "Metrics"
// @Success 200 {array} view.Metric
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Error"
// @Router /updates [post]
// Конец Swagger описания.
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
