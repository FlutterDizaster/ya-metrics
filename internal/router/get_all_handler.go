package router

import (
	"html/template"
	"net/http"
)

func (r *Router) getAllHandler(w http.ResponseWriter, _ *http.Request) {
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
		http.Error(w, "Error whlie creating temaplate", http.StatusInternalServerError)
		return
	}

	metrics := r.storage.ReadAllMetrics()

	err = tmpl.ExecuteTemplate(w, "metrics", metrics)

	if err != nil {
		http.Error(w, "Error whlie executing temaplate", http.StatusInternalServerError)
		return
	}
}