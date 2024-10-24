package main

import (
	"embed"
	"fmt"
	"html/template"
	iofs "io/fs"
	"log"
	"net/http"
	"os"

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

// inspiration see: https://forum.golangbridge.org/t/can-i-use-enum-in-template/25296
var funcMap = template.FuncMap{
	"IsMatchVague": puzzle.MatchVague.Is,
	"IsMatchNone":  puzzle.MatchNone.Is,
	"IsMatchExact": puzzle.MatchExact.Is,
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
	mux.HandleFunc("GET /", routes.Index(t, &sessions, wordDb, envCfg.imprintUrl, Revision, FaviconPath))
	mux.HandleFunc("GET /test", routes.TestPage(t))
	mux.HandleFunc("GET /letter-hint", routes.LetterHint(t, &sessions, wordDb))
	mux.HandleFunc("GET /lettr", routes.GetLettr(t, &sessions, wordDb, envCfg.imprintUrl, Revision, FaviconPath))
	mux.HandleFunc("POST /lettr", routes.PostLettr(t, &sessions, wordDb, envCfg.imprintUrl, Revision, FaviconPath))
	mux.HandleFunc("POST /new", routes.PostNew(t, &sessions, wordDb, envCfg.imprintUrl, Revision, FaviconPath))
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
