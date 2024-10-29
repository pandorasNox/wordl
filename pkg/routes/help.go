package routes

import (
	"log"
	"net/http"

	"github.com/pandorasNox/lettr/pkg/puzzle"
	"github.com/pandorasNox/lettr/pkg/routes/models"
	"github.com/pandorasNox/lettr/pkg/session"
)

func Help(sessions *session.Sessions, wdb puzzle.WordDatabase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := session.HandleSession(w, r, sessions, wdb)
		g := s.GameState()
		sessions.UpdateOrSet(s)

		td := models.TemplateDataHelpPage{
			SolutionWord:                g.ActiveSolutionWord().String(),
			PastWords:                   s.PastWords(),
			LetterHints:                 g.LetterHints(),
			SolutionHasDublicateLetters: g.ActiveSolutionWord().HasDublicateLetters(),
		}

		err := routesTemplates.ExecuteTemplate(w, "help", td)
		if err != nil {
			log.Printf("error t.ExecuteTemplate '/help' route: %s", err)
		}
	}
}
