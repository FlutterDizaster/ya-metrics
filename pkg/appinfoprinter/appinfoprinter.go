package appinfoprinter

import (
	"html/template"
	"os"
)

type AppInfo struct {
	Version string
	Date    string
	Commit  string
}

const appInfoTemplate = `Build version: {{if .Version}}{{.Version}}{{else}}N/A{{end}}
Build date: {{if .Date}}{{.Date}}{{else}}N/A{{end}}
Build commit: {{if .Commit}}{{.Commit}}{{else}}N/A{{end}}
`

func PrintAppInfo(info AppInfo) error {
	tmpl := template.Must(template.New("appinfo").Parse(appInfoTemplate))

	return tmpl.Execute(os.Stdout, info)
}
