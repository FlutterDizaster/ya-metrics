package appinfoprinter

import (
	"embed"
	"html/template"
	"os"
)

type AppInfo struct {
	Version string
	Date    string
	Commit  string
}

//go:embed templates/appinfo.tmpl
var appInfoTemplateFS embed.FS

func PrintAppInfo(info AppInfo) error {
	appInfoTemplateContent, err := appInfoTemplateFS.ReadFile("templates/appinfo.tmpl")
	if err != nil {
		return err
	}

	appInfoTemplate := template.Must(template.New("appinfo").Parse(string(appInfoTemplateContent)))

	return appInfoTemplate.Execute(os.Stdout, info)
}
