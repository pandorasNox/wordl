package routes

import "embed"

//go:embed templates/*.html.tmpl
//go:embed templates/**/*.html.tmpl
var TemplatesFs embed.FS
