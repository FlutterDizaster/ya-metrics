package api

import (
	"bytes"

	"fmt"
	"log/slog"
	"net/http"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
	"github.com/go-chi/chi/v5"
)

// Handler для обноваления состояния метрики в репозитории.
//
// Swagger описание:
// @Summary Update metric
// @Description Update metric in DB
// @Tags metrics
// @Produce text/plain
// @Param kind path string true "Metric kind"
// @Param name path string true "Metric name"
// @Param value path string true "Metric value"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Error"
// @Router /update/{kind}/{name}/{value} [post]
// Конец Swagger описания.
func (api *API) updateHandler(w http.ResponseWriter, req *http.Request) {
	// парсинг URL запроса
	kind := chi.URLParam(req, "kind")
	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")

	// создание новой метрики
	metric, err := view.NewMetric(kind, name, value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// добавление метрики в репозиторий
	if _, err = api.storage.AddMetrics(*metric); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// записываем ответ
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Add("Content-Type", "charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// Метод обрабатывает POST-запросы на добавление метрики в репозиторий.
//
// Swagger описание:
// @Summary Update metric
// @Description Update metric in DB in JSON format
// @Tags metrics
// @Produce json
// @Param metric body view.Metric true "Metric"
// @Success 200 {object} view.Metric
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Error"
// @Router /update [post]
// Конец Swagger описания.
func (api *API) updateJSONHandler(w http.ResponseWriter, req *http.Request) {
	var metric view.Metric
	var buf bytes.Buffer

	// читаем тело запроса
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Unmarshal тела запроса
	if err = metric.UnmarshalJSON(buf.Bytes()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// добавление метрики в репозиторий
	metrics, err := api.storage.AddMetrics(metric)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Marshal ответа
	resp, err := metrics[0].MarshalJSON()
	if err != nil {
		slog.Error(
			"marshaling error",
			"message", err,
			"metric", metric.ID,
			"value", metric.StringValue(),
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
