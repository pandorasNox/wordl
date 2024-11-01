package routes

import (
	"log"
	"net/http"

	"github.com/pandorasNox/lettr/pkg/router/routes/templates"
)

type TemplateDataTestPage struct{}

func TestPage() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		tpDate := TemplateDataTestPage{}

		err := templates.Routes.ExecuteTemplate(w, "test.html.tmpl", tpDate)
		if err != nil {
			log.Printf("error t.Execute '/' route: %s", err)
		}
	}
}
