package api

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"sort"

	"github.com/FlutterDizaster/ya-metrics/internal/view"
)

// Handler отдающий таблицу со всеми имеющимися метриками и их значениями.
func (api *API) getAllHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка на принимаемы Content-Type
	if r.Header.Get("Accept") == "application/json" {
		api.getAllJSONHandler(w, r)
		return
	}
	// шаблон html страницы с ответом
	// TODO: вынести в отдельный template файл
	content := `{{define "metrics"}}
	<!doctype html>
	<html lang="en">
		<head>
			<title>Metrics</title>
		</head>
		<body>
			<table>
				<th>Kind</th>
				<th>Name</th>
				<th>Value</th>
				{{range .}}
					<tr>
						<td>{{.MType}}</td>
						<td>{{.ID}}</td>
						{{if eq .MType "gauge"}}
							<td>{{.Value}}</td>
						{{else if eq .MType "counter"}}
							<td>{{.Delta}}</td>
						{{end}}
					</tr>
				{{end}}
			</table>
		</body>
	</html>
	{{end}}`

	// парсинг темплейта
	tmpl, err := template.New("metrics").Parse(content)
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
	if _, err = w.Write(resp.Bytes()); err != nil {
		slog.Error("writing response error", "message", err)
		http.Error(w, fmt.Sprintf("write metric error: %s", err), http.StatusInternalServerError)
		return
	}
}

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
