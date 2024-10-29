package routes

import (
	"log"
	"net/http"
)

type TemplateDataTestPage struct{}

func TestPage() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		tpDate := TemplateDataTestPage{}

		err := routesTemplates.ExecuteTemplate(w, "test.html.tmpl", tpDate)
		if err != nil {
			log.Printf("error t.Execute '/' route: %s", err)
		}
	}
}
