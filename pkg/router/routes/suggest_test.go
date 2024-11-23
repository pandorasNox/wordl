package routes

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/pandorasNox/lettr/pkg/language"
	"github.com/pandorasNox/lettr/pkg/puzzle"
	"github.com/pandorasNox/lettr/pkg/session"
)

func TestGetSuggest(t *testing.T) {
	mockFs := fstest.MapFS{
		"all.txt": {
			// Data: []byte("hello, world"),
			Data: []byte(`# metadata
gamer
games
`),
		},
		"common.txt": {
			Data: []byte(`# metadata
cried
`),
		},
	}

	fMap := map[language.Language]map[puzzle.WordCollection][]string{
		language.LANG_EN: {
			puzzle.WC_ALL: {
				"all.txt",
			},
			puzzle.WC_COMMON: {
				"common.txt",
			},
		},
	}

	wordDb := puzzle.WordDatabase{}
	err := wordDb.Init(mockFs, fMap)
	if err != nil {
		log.Fatalf("init wordDatabase failed: %s", err)
	}

	sessions := session.NewSessions()

	getSuggestHandler := GetSuggest(&sessions, wordDb)

	req := httptest.NewRequest(http.MethodGet, "/lettr", nil)
	recorder := httptest.NewRecorder()

	getSuggestHandler(recorder, req)

	// extract cookie
	cookie := recorder.Result().Cookies()[0]

	sess, err := sessions.GetById(cookie.Value)
	if err != nil {
		t.Errorf("couldn't get session by id='%s', error: %s", cookie.Value, err)
	}

	if sess.SecurityHoneypotMessageInputName() == "" {
		t.Errorf("expected sess.SecurityHoneypotMessageInputName() to not be emty string, got '%v'", sess.SecurityHoneypotMessageInputName())
	}
}
