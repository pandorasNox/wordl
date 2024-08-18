package routes

import (
	"html/template"
	"net/http"
)

type TemplateDataLetterHint struct{}

func LetterHint(t *template.Template) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {

		w.Write([]byte("H"))
	}
}
