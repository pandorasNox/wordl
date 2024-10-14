package main

import (
	"embed"
	"fmt"
	"html/template"
	iofs "io/fs"
	"log"
	"net/http"
	"os"
	"time"
	"unicode"

	"github.com/pandorasNox/lettr/pkg/language"
	"github.com/pandorasNox/lettr/pkg/middleware"
	"github.com/pandorasNox/lettr/pkg/puzzle"
	"github.com/pandorasNox/lettr/pkg/routes"
	"github.com/pandorasNox/lettr/pkg/session"
)

var Revision = "0000000"
var FaviconPath = "/static/assets/favicon"

//go:embed configs/*.txt
//go:embed templates/*.html.tmpl
//go:embed templates/**/*.html.tmpl
//go:embed web/static/assets/*
//go:embed web/static/generated/*.js
//go:embed web/static/generated/*.css
var fs embed.FS

type env struct {
	port        string
	githubToken string
	imprintUrl  string
}

func (e env) String() string {
	s := fmt.Sprintf("port: %s", e.port)

	if e.githubToken != "" {
		s = fmt.Sprintf("%s\ngithub token (length): %d", s, len(e.githubToken))
	}

	if e.imprintUrl != "" {
		s = fmt.Sprintf("%s\nimprint: %s", s, e.imprintUrl)
	}
	// s = s + fmt.Sprintf("foo: %s\n", e.port)
	return s
}

func NewLang(maybeLang string) (language.Language, error) {
	switch language.Language(maybeLang) {
	case language.LANG_EN, language.LANG_DE:
		return language.Language(maybeLang), nil
	default:
		return language.LANG_EN, fmt.Errorf("couldn't create new language from given value: '%s'", maybeLang)
	}
}

// inspiration see: https://forum.golangbridge.org/t/can-i-use-enum-in-template/25296
var funcMap = template.FuncMap{
	"IsMatchVague": puzzle.MatchVague.Is,
	"IsMatchNone":  puzzle.MatchNone.Is,
	"IsMatchExact": puzzle.MatchExact.Is,
}

type TemplateDataForm struct {
	Data                        puzzle.Puzzle
	Errors                      map[string]string
	IsSolved                    bool
	IsLoose                     bool
	JSCachePurgeTimestamp       int64
	Language                    language.Language
	Revision                    string
	FaviconPath                 string
	Keyboard                    keyboard
	PastWords                   []puzzle.Word
	SolutionHasDublicateLetters bool
	ImprintUrl                  string
}

func (fd TemplateDataForm) New(l language.Language, p puzzle.Puzzle, pastWords []puzzle.Word, solutionHasDublicateLetters bool, imprintUrl string) TemplateDataForm {
	kb := keyboard{}
	kb.Init(l, p.LetterGuesses())

	return TemplateDataForm{
		Data:                        p,
		Errors:                      make(map[string]string),
		JSCachePurgeTimestamp:       time.Now().Unix(),
		Language:                    l,
		Revision:                    Revision,
		FaviconPath:                 FaviconPath,
		Keyboard:                    kb,
		PastWords:                   pastWords,
		SolutionHasDublicateLetters: solutionHasDublicateLetters,
		ImprintUrl:                  imprintUrl,
	}
}

type keyboard struct {
	KeyGrid [][]keyboardKey
}

func (k *keyboard) Init(l language.Language, lgs []puzzle.LetterGuess) {
	k.KeyGrid = [][]keyboardKey{
		{{"Q", false, puzzle.MatchNone}, {"W", false, puzzle.MatchNone}, {"E", false, puzzle.MatchNone}, {"R", false, puzzle.MatchNone}, {"T", false, puzzle.MatchNone}, {"Y", false, puzzle.MatchNone}, {"U", false, puzzle.MatchNone}, {"I", false, puzzle.MatchNone}, {"O", false, puzzle.MatchNone}, {"P", false, puzzle.MatchNone}, {"Delete", false, puzzle.MatchNone}},
		{{"A", false, puzzle.MatchNone}, {"S", false, puzzle.MatchNone}, {"D", false, puzzle.MatchNone}, {"F", false, puzzle.MatchNone}, {"G", false, puzzle.MatchNone}, {"H", false, puzzle.MatchNone}, {"J", false, puzzle.MatchNone}, {"K", false, puzzle.MatchNone}, {"L", false, puzzle.MatchNone}, {"Enter", false, puzzle.MatchNone}},
		{{"Z", false, puzzle.MatchNone}, {"X", false, puzzle.MatchNone}, {"C", false, puzzle.MatchNone}, {"V", false, puzzle.MatchNone}, {"B", false, puzzle.MatchNone}, {"N", false, puzzle.MatchNone}, {"M", false, puzzle.MatchNone}},
	}

	for ri, r := range k.KeyGrid {
	KeyLoop:
		for ki, kk := range r {
			for _, lg := range lgs {
				if kk.Key == "Enter" || kk.Key == "Delete" {
					continue KeyLoop
				}

				KeyR := firstRune(kk.Key)
				betterMatch := (k.KeyGrid[ri][ki].Match == puzzle.MatchNone) ||
					(k.KeyGrid[ri][ki].Match == puzzle.MatchVague && lg.Match == puzzle.MatchExact)

				if lg.Letter == unicode.ToLower(KeyR) && betterMatch {
					k.KeyGrid[ri][ki].IsUsed = true
					k.KeyGrid[ri][ki].Match = lg.Match
				}
			}
		}
	}
}

func firstRune(s string) rune {
	for _, r := range s {
		return r
	}

	return 0
}

type keyboardKey struct {
	Key    string
	IsUsed bool
	Match  puzzle.Match
}

func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}

func main() {
	log.Println("staring server...")

	envCfg := envConfig()
	sessions := session.NewSessions()

	wordDb := puzzle.WordDatabase{}
	err := wordDb.Init(fs, puzzle.FilePathsByLang())
	if err != nil {
		log.Fatalf("init wordDatabase failed: %s", err)
	}

	log.Printf("env conf:\n%s", envCfg)

	// t := template.Must(template.ParseFS(fs, "templates/index.html.tmpl", "templates/lettr-form.html.tmpl"))
	// log.Printf("template name: %s", t.Name())
	t := template.Must(template.New("index.html.tmpl").Funcs(funcMap).ParseFS(
		fs,
		"templates/index.html.tmpl",
		"templates/lettr-form.html.tmpl",
		"templates/help.html.tmpl",
		"templates/suggest.html.tmpl",
		"templates/pages/test.html.tmpl",
	))

	staticFS, err := iofs.Sub(fs, "web/static")
	if err != nil {
		log.Fatalf("subtree for 'static' dir of embed fs failed: %s", err) //TODO
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /static/", routes.Static(staticFS))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, req *http.Request) {
		sess := session.HandleSession(w, req, &sessions, wordDb)

		p := sess.GameState().LastEvaluatedAttempt()
		sessions.UpdateOrSet(sess)

		fData := TemplateDataForm{}.New(sess.Language(), p, sess.PastWords(), sess.GameState().ActiveSolutionWord().HasDublicateLetters(), envCfg.imprintUrl)
		fData.IsSolved = p.IsSolved()
		fData.IsLoose = p.IsLoose()

		err := t.ExecuteTemplate(w, "index.html.tmpl", fData)
		if err != nil {
			log.Printf("error t.Execute '/' route: %s", err)
		}
	})

	mux.HandleFunc("GET /test", routes.TestPage(t))

	mux.HandleFunc("GET /letter-hint", routes.LetterHint(t, &sessions, wordDb))

	mux.HandleFunc("GET /lettr", routes.GetLettr(t, &sessions, wordDb, envCfg.imprintUrl, Revision, FaviconPath))

	mux.HandleFunc("POST /lettr", routes.PostLettr(t, &sessions, wordDb, envCfg.imprintUrl, Revision, FaviconPath))

	mux.HandleFunc("POST /new", func(w http.ResponseWriter, r *http.Request) {
		s := session.HandleSession(w, r, &sessions, wordDb)

		// handle lang switch
		l := s.Language()
		maybeLang := r.FormValue("lang")
		if maybeLang != "" {
			l, _ = NewLang(maybeLang)
			s.SetLanguage(l)

			type TemplateDataLanguge struct {
				Language language.Language
			}
			tData := TemplateDataLanguge{Language: l}

			err := t.ExecuteTemplate(w, "oob-lang-switch", tData)
			if err != nil {
				log.Printf("error t.ExecuteTemplate '/new' route: %s", err)
			}
		}

		p := puzzle.Puzzle{}

		s.AddPastWord(s.GameState().ActiveSolutionWord())
		s.NewGame(l, wordDb)
		sessions.UpdateOrSet(s)

		g := s.GameState()
		fData := TemplateDataForm{}.New(s.Language(), p, s.PastWords(), g.ActiveSolutionWord().HasDublicateLetters(), envCfg.imprintUrl)
		fData.IsSolved = p.IsSolved()
		fData.IsLoose = p.IsLoose()

		// w.Header().Add("HX-Refresh", "true")
		err := t.ExecuteTemplate(w, "lettr-form", fData)
		if err != nil {
			log.Printf("error t.ExecuteTemplate '/new' route: %s", err)
		}
	})

	mux.HandleFunc("POST /help", routes.Help(t, &sessions, wordDb))

	mux.HandleFunc("GET /suggest", routes.GetSuggest(t))

	mux.HandleFunc("POST /suggest", routes.PostSuggest(t, envCfg.githubToken))

	middlewares := []func(h http.Handler) http.Handler{
		func(h http.Handler) http.Handler {
			return middleware.NewRequestSize(h, 32*1024 /* 32kiB */)
		},
		func(h http.Handler) http.Handler {
			return middleware.NewBodySize(h, 32*1024 /* 32kiB */)
		},
	}

	var muxWithMiddlewares http.Handler = mux
	for _, fm := range middlewares {
		muxWithMiddlewares = fm(muxWithMiddlewares)
	}

	// v1 := http.NewServeMux()
	// v1.Handle("/v1/", http.StripPrefix("/v1", muxWithMiddlewares))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", envCfg.port), muxWithMiddlewares))
}

func envConfig() env {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		panic("PORT not provided")
	}

	gt, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		log.Printf("(optional) environment variable GITHUB_TOKEN not set")
	}

	imprintUrl, ok := os.LookupEnv("IMPRINT_URL")
	if !ok {
		log.Printf("(optional) environment variable IMPRINT_URL not set")
	}

	return env{port: port, githubToken: gt, imprintUrl: imprintUrl}
}
