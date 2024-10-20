package routes

import (
	"html/template"
	"log"
	"net/http"

	"github.com/pandorasNox/lettr/pkg/puzzle"
	"github.com/pandorasNox/lettr/pkg/routes/models"
	"github.com/pandorasNox/lettr/pkg/session"
)

func Index(t *template.Template, sessions *session.Sessions, wdb puzzle.WordDatabase, imprintUrl string, revision string, faviconPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := session.HandleSession(w, r, sessions, wdb)

		p := sess.GameState().LastEvaluatedAttempt()
		sessions.UpdateOrSet(sess)

		fData := models.TemplateDataIndex{}.New(sess.Language(), p, sess.PastWords(), imprintUrl, revision, faviconPath)
		fData.IsSolved = p.IsSolved()
		fData.IsLoose = p.IsLoose()

		err := t.ExecuteTemplate(w, "index.html.tmpl", fData)
		if err != nil {
			log.Printf("error t.Execute '/' route: %s", err)
		}
	}
}
