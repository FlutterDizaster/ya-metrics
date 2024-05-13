package api

import (
	"html/template"
	"net/http"
	"sort"
)

// Handler отдающий таблицу со всеми имеющимися метриками и их значениями.
func (r *Router) getAllHandler(w http.ResponseWriter, _ *http.Request) {
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
	metrics := r.storage.ReadAllMetrics()
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})

	// передача ответа клиенту
	err = tmpl.ExecuteTemplate(w, "metrics", metrics)
	if err != nil {
		http.Error(w, "Error whlie executing temaplate", http.StatusInternalServerError)
		return
	}
}
