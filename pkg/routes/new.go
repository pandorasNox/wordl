package routes

import (
	"log"
	"net/http"

	"github.com/pandorasNox/lettr/pkg/language"
	"github.com/pandorasNox/lettr/pkg/puzzle"
	"github.com/pandorasNox/lettr/pkg/routes/models/shared"
	"github.com/pandorasNox/lettr/pkg/session"
)

func PostNew(sessions *session.Sessions, wdb puzzle.WordDatabase, imprintUrl string, revision string, faviconPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := session.HandleSession(w, r, sessions, wdb)

		// handle lang switch
		l := s.Language()
		maybeLang := r.FormValue("lang")
		if maybeLang != "" {
			l, _ = language.NewLang(maybeLang)
			s.SetLanguage(l)

			type TemplateDataLanguge struct {
				Language language.Language
			}
			tData := TemplateDataLanguge{Language: l}

			err := routesTemplates.ExecuteTemplate(w, "oob-lang-switch", tData)
			if err != nil {
				log.Printf("error t.ExecuteTemplate '/new' route: %s", err)
			}
		}

		p := puzzle.Puzzle{}

		s.AddPastWord(s.GameState().ActiveSolutionWord())
		s.NewGame(l, wdb)
		sessions.UpdateOrSet(s)

		fData := shared.TemplateDataLettr{}.New(s.Language(), p, s.PastWords(), imprintUrl, revision, faviconPath)
		fData.IsSolved = p.IsSolved()
		fData.IsLoose = p.IsLoose()

		// w.Header().Add("HX-Refresh", "true")
		err := routesTemplates.ExecuteTemplate(w, "lettr-form", fData)
		if err != nil {
			log.Printf("error t.ExecuteTemplate '/new' route: %s", err)
		}
	}
}
