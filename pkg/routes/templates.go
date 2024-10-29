package routes

import (
	"embed"
	"html/template"

	"github.com/pandorasNox/lettr/pkg/puzzle"
)

//go:embed templates/*.html.tmpl
//go:embed templates/**/*.html.tmpl
var templatesFs embed.FS

// inspiration see: https://forum.golangbridge.org/t/can-i-use-enum-in-template/25296
var funcMap = template.FuncMap{
	"IsMatchVague": puzzle.MatchVague.Is,
	"IsMatchNone":  puzzle.MatchNone.Is,
	"IsMatchExact": puzzle.MatchExact.Is,
}

// routesTemplate := template.Must(template.ParseFS(fs, "routesTemplates/index.html.tmpl", "routesTemplates/lettr-form.html.tmpl"))
// log.Printf("template name: %s", routesTemplate.Name())
var routesTemplates = template.Must(template.New("index.html.tmpl").Funcs(funcMap).ParseFS(
	templatesFs,
	"templates/index.html.tmpl",
	"templates/lettr-form.html.tmpl",
	"templates/help.html.tmpl",
	"templates/suggest.html.tmpl",
	"templates/pages/test.html.tmpl",
))
