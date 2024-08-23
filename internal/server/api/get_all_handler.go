package api

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"sort"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

//go:embed templates/metrics.html
var tmpl string

// Handler отдающий таблицу со всеми имеющимися метриками и их значениями.
//
// Swagger описание:
// @Summary Get all metrics
// @Description Get all metrics
// @Tags metrics
// @Produce html/json
// @Success 200 {array} view.Metric
// @Failure 500 {string} string "Error"
// @Router / [get]
// Конец Swagger описания.
func (api *API) getAllHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка на принимаемы Content-Type
	if r.Header.Get("Accept") == "application/json" {
		api.getAllJSONHandler(w, r)
		return
	}

	// парсинг темплейта
	tmpl, err := template.New("metrics").Parse(tmpl)
	if err != nil {
		http.Error(w, "Error whlie creating temaplate", http.StatusInternalServerError)
		return
	}

	// получение всех метрик из репозитория
	metrics, err := api.storage.ReadAllMetrics()
	if err != nil {
		http.Error(w, "Error whlie getting metrics from repository", http.StatusInternalServerError)
		return
	}
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})

	// Компиляция темплейта
	var resp bytes.Buffer
	err = tmpl.ExecuteTemplate(&resp, "metrics", metrics)
	if err != nil {
		http.Error(w, "Error whlie executing temaplate", http.StatusInternalServerError)
		return
	}

	// Передача ответа клиенту
	w.Header().Set("Content-Type", "text/html")
	if _, err = w.Write(resp.Bytes()); err != nil {
		slog.Error("writing response error", "message", err)
		http.Error(w, fmt.Sprintf("write metric error: %s", err), http.StatusInternalServerError)
		return
	}
}

// getAllJSONHandler обрабатывает GET-запросы на получение таблицы со всеми имеющимися метриками в формате JSON.
func (api *API) getAllJSONHandler(w http.ResponseWriter, _ *http.Request) {
	metrics, err := api.storage.ReadAllMetrics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})

	resp, err := view.Metrics(metrics).MarshalJSON()
	if err != nil {
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
