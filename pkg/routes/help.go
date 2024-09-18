package routes

import (
	"html/template"
	"log"
	"net/http"

	"github.com/pandorasNox/lettr/pkg/puzzle"
	"github.com/pandorasNox/lettr/pkg/session"
)

type TemplateDataHelpPage struct {
	SolutionWord                string
	SolutionHasDublicateLetters bool
	LetterHints                 string
	PastWords                   []puzzle.Word
}

func Help(t *template.Template, sessions *session.Sessions, wdb puzzle.WordDatabase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := session.HandleSession(w, r, sessions, wdb)
		sessions.UpdateOrSet(s)

		td := TemplateDataHelpPage{
			SolutionWord:                s.ActiveSolutionWord().String(),
			PastWords:                   s.PastWords(),
			LetterHints:                 string(s.LetterHints()),
			SolutionHasDublicateLetters: s.ActiveSolutionWord().HasDublicateLetters(),
		}

		err := t.ExecuteTemplate(w, "help", td)
		if err != nil {
			log.Printf("error t.ExecuteTemplate '/help' route: %s", err)
		}
	}
}
