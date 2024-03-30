package handlers

import (
	"net/http"
	"text/template"
)

type GetAllHandler struct {
	storage GetAllMetricsStorage
}

func NewGetAllHandler(storage GetAllMetricsStorage) GetAllHandler {
	return GetAllHandler{
		storage: storage,
	}
}

func (h GetAllHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
						<td>{{.Kind}}</td>
						<td>{{.Name}}</td>
						<td>{{.Value}}</td>
					</tr>
				{{end}}
			</table>
		</body>
	</html>
	{{end}}`

	tmpl, err := template.New("metrics").Parse(content)

	if err != nil {
		http.Error(w, "Error whlie creating temaplate", 500)
		return
	}

	metrics := h.storage.ReadAllMetrics()

	err = tmpl.ExecuteTemplate(w, "metrics", metrics)

	if err != nil {
		http.Error(w, "Error whlie executing temaplate", 500)
		return
	}
}
