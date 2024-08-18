package main

import (
	"bufio"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"unicode"

	iofs "io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/pandorasNox/lettr/pkg/github"
	"github.com/pandorasNox/lettr/pkg/middleware"
	"github.com/pandorasNox/lettr/pkg/puzzle"
	"github.com/pandorasNox/lettr/pkg/routes"
)

var Revision = "0000000"
var FaviconPath = "/static/assets/favicon"

const SESSION_COOKIE_NAME = "session"
const SESSION_MAX_AGE_IN_SECONDS = 24 * 60 * 60

//go:embed configs/*.txt
//go:embed templates/*.html.tmpl
//go:embed templates/**/*.html.tmpl
//go:embed web/static/assets/*
//go:embed web/static/generated/*.js
//go:embed web/static/generated/*.css
var fs embed.FS

var ErrNotInWordList = errors.New("not in wordlist")

type env struct {
	port        string
	githubToken string
}

func (e env) String() string {
	s := fmt.Sprintf("port: %s\ngithub token (length): %d\n", e.port, len(e.githubToken))
	// s = s + fmt.Sprintf("foo: %s\n", e.port)
	return s
}

type counterState struct {
	mu    sync.Mutex
	count int
}

type session struct {
	// todo: think about using mutex or channel for rw session
	id                   string
	expiresAt            time.Time
	maxAgeSeconds        int
	language             language
	activeSolutionWord   puzzle.Word
	lastEvaluatedAttempt puzzle.Puzzle
	pastWords            []puzzle.Word
}

func (s *session) AddPastWord(w puzzle.Word) {
	s.pastWords = append(s.pastWords, w)
}

func (s *session) PastWords() []puzzle.Word {
	return slices.Clone(s.pastWords)
}

type sessions []session

func (ss sessions) String() string {
	out := ""
	for _, s := range ss {
		out = out + s.id + " " + s.expiresAt.String() + "\n"
	}

	return out
}

func (ss *sessions) updateOrSet(sess session) {
	index := slices.IndexFunc((*ss), func(s session) bool {
		return s.id == sess.id
	})
	if index == -1 {
		*ss = append(*ss, sess)
		return
	}

	(*ss)[index] = sess
}

type language string

func NewLang(maybeLang string) (language, error) {
	switch language(maybeLang) {
	case LANG_EN, LANG_DE:
		return language(maybeLang), nil
	default:
		return LANG_EN, fmt.Errorf("couldn't create new language from given value: '%s'", maybeLang)
	}
}

const (
	LANG_EN language = "en"
	LANG_DE language = "de"
)

func toWord(wo string) (puzzle.Word, error) {
	out := puzzle.Word{}

	length := 0
	for i, l := range wo {
		length++
		if length > len(out) {
			return puzzle.Word{}, fmt.Errorf("string does not match allowed word length: length=%d, expectedLength=%d", length, len(out))
		}

		out[i] = l
	}

	if length < len(out) {
		return puzzle.Word{}, fmt.Errorf("string is to short: length=%d, expectedLength=%d", length, len(out))
	}

	return out, nil
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
	Language                    language
	Revision                    string
	FaviconPath                 string
	Keyboard                    keyboard
	PastWords                   []puzzle.Word
	SolutionHasDublicateLetters bool
}

func (fd TemplateDataForm) New(l language, p puzzle.Puzzle, pastWords []puzzle.Word, SolutionHasDublicateLetters bool) TemplateDataForm {
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
		SolutionHasDublicateLetters: SolutionHasDublicateLetters,
	}
}

type TemplateDataSuggest struct {
	Word     string
	Message  string
	Language string
	Action   string
}

var RegexpAllowedWordCharacters = regexp.MustCompile(`^[A-Za-zöäüÖÄÜß]{5}$`)

func (tds TemplateDataSuggest) validate() error {
	if !RegexpAllowedWordCharacters.Match([]byte(tds.Word)) {
		return fmt.Errorf("validation failed: %s", "word is either to long, to short or contains forbidden characters")
	}

	p := bluemonday.UGCPolicy()
	sm := p.Sanitize(tds.Message)
	if sm != tds.Message {
		return fmt.Errorf("validation failed: %s", "message contains invalid data")
	}

	if !slices.Contains([]string{"add", "remove"}, tds.Action) {
		return fmt.Errorf("validation failed: %s", "action invalid")
	}

	if tds.Language != "german" && tds.Language != "english" {
		return fmt.Errorf("validation failed: %s", "language invalid")
	}

	return nil
}

type wordCollection string

const (
	WC_ALL    wordCollection = "wc_all"
	WC_COMMON wordCollection = "wc_common"
)

type wordDatabase struct {
	db map[language]map[wordCollection]map[puzzle.Word]bool
}

func (wdb *wordDatabase) Init(fs iofs.FS, filePathsByLanguage map[language]map[wordCollection][]string) error {
	wdb.db = make(map[language]map[wordCollection]map[puzzle.Word]bool)

	for l, collection := range filePathsByLanguage {
		wdb.db[l] = make(map[wordCollection]map[puzzle.Word]bool)
		for c, paths := range collection {
			wdb.db[l][c] = make(map[puzzle.Word]bool)

			for _, path := range paths {
				f, err := fs.Open(path)
				if err != nil {
					return fmt.Errorf("wordDatabase init failed when opening file: %s", err)
				}
				defer f.Close()

				fInfo, err := f.Stat()
				if err != nil {
					return fmt.Errorf("wordDatabase init failed when obtaining stat: %s", err)
				}

				var allowedSize int64 = 2 * 1024 * 1024 // 2 MB
				if fInfo.Size() > allowedSize {
					return fmt.Errorf("wordDatabase init failed with forbidden file size: path='%s', size='%d'", path, fInfo.Size())
				}

				scanner := bufio.NewScanner(f)
				var line int = 0
				for scanner.Scan() {
					if line == 0 { // skip first metadata line
						line++
						continue
					}

					candidate := scanner.Text()
					word, err := toWord(candidate)
					if err != nil {
						return fmt.Errorf("wordDatabase init, couldn't parse line to word: line='%s', err=%s", candidate, err)
					}

					wdb.db[l][c][word.ToLower()] = true

					line++
				}
				if err := scanner.Err(); err != nil {
					return fmt.Errorf("wordDatabase init failed scanning file with: path='%s', err=%s", path, err)
				}
			}
		}
	}

	return nil
}

func (wdb wordDatabase) Exists(l language, w puzzle.Word) bool {
	db, ok := wdb.db[l]
	if !ok {
		return false
	}

	db_c, ok := db[WC_ALL]
	if !ok {
		return false
	}

	_, ok = db_c[w.ToLower()]
	return ok
}

func (wdb wordDatabase) RandomPick(l language, avoidList []puzzle.Word, retryAkkumulator uint8) (puzzle.Word, error) {
	const MAX_RETRY uint8 = 10

	if retryAkkumulator > MAX_RETRY {
		return puzzle.Word{}, fmt.Errorf("RandomPick exceeded retries: retryAkkumulator='%d' | MAX_RETRY='%d'", retryAkkumulator, MAX_RETRY)
	}

	db, ok := wdb.db[l]
	if !ok {
		return puzzle.Word{}, fmt.Errorf("RandomPick failed with unknown language: '%s'", l)
	}

	collection := WC_COMMON
	db_c, ok := db[collection]
	if !ok {
		collection = WC_ALL

		db_c, ok = db[collection]
		if !ok {
			return puzzle.Word{}, fmt.Errorf("RandomPick with lang '%s' failed with unknown collection: '%s'", l, collection)
		}
	}

	randsource := rand.NewSource(time.Now().UnixNano())
	randgenerator := rand.New(randsource)
	rolledLine := randgenerator.Intn(len(db_c))

	currentLine := 0
	for w := range db_c {
		if currentLine == rolledLine {

			wordContained := slices.ContainsFunc(avoidList, func(wo puzzle.Word) bool {
				return w.IsEqual(wo)
			})
			if wordContained {
				return wdb.RandomPick(l, avoidList, retryAkkumulator+1)
			}

			return w, nil
		}

		currentLine++
	}

	return puzzle.Word{}, fmt.Errorf("RandomPick could not find random line aka this should never happen ^^")
}

func (wdb wordDatabase) RandomPickWithFallback(l language, avoidList []puzzle.Word, retryAkkumulator uint8) puzzle.Word {
	w, err := wdb.RandomPick(l, avoidList, retryAkkumulator)
	if err != nil {
		return puzzle.Word{'R', 'O', 'A', 'T', 'E'}.ToLower()
	}

	return w.ToLower()
}

type keyboard struct {
	KeyGrid [][]keyboardKey
}

func (k *keyboard) Init(l language, lgs []puzzle.LetterGuess) {
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

type Message string

type TemplateDataMessages struct {
	ErrMsgs     []Message
	InfoMsgs    []Message
	SuccessMsgs []Message
}

func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}

func getAllFilenames(efs iofs.FS) (files []string, err error) {
	if err := iofs.WalkDir(efs, ".", func(path string, d iofs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		files = append(files, path)

		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

func filePathsByLang() map[language]map[wordCollection][]string {
	return map[language]map[wordCollection][]string{
		LANG_EN: {
			WC_ALL: {
				"configs/corpora-eng_news_2023_10K-export.txt",
				"configs/en-en.words.v2.txt",
			},
			WC_COMMON: {
				"configs/corpora-eng_news_2023_10K-export.txt",
			},
		},
		LANG_DE: {
			WC_ALL: {
				"configs/corpora-deu_news_2023_10K-export.txt",
				"configs/de-de.words.v2.txt",
			},
			WC_COMMON: {
				"configs/corpora-deu_news_2023_10K-export.txt",
			},
		},
	}
}

func main() {
	log.Println("staring server...")

	envCfg := envConfig()
	sessions := sessions{}

	wordDb := wordDatabase{}
	err := wordDb.Init(fs, filePathsByLang())
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

	mux := http.NewServeMux()

	staticFS, err := iofs.Sub(fs, "web/static")
	if err != nil {
		log.Fatalf("subtree for 'static' dir of embed fs failed: %s", err)
	}

	mux.Handle(
		"GET /static/",
		http.StripPrefix("/static", http.FileServer(http.FS(staticFS))),
	)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, req *http.Request) {
		sess := handleSession(w, req, &sessions, wordDb)

		p := sess.lastEvaluatedAttempt
		// log.Printf("debug '/' route - sess.lastEvaluatedAttempt:\n %v\n", wo)
		p.Debug = sess.activeSolutionWord.String()
		sessions.updateOrSet(sess)

		fData := TemplateDataForm{}.New(sess.language, p, sess.PastWords(), sess.activeSolutionWord.HasDublicateLetters())
		fData.IsSolved = p.IsSolved()
		fData.IsLoose = p.IsLoose()

		err := t.ExecuteTemplate(w, "index.html.tmpl", fData)
		if err != nil {
			log.Printf("error t.Execute '/' route: %s", err)
		}
	})

	mux.HandleFunc("GET /test", routes.TestPage(t))
	mux.HandleFunc("GET /letter-hint", routes.LetterHint(t))

	mux.HandleFunc("GET /lettr", func(w http.ResponseWriter, r *http.Request) {
		s := handleSession(w, r, &sessions, wordDb)

		p := s.lastEvaluatedAttempt

		sessions.updateOrSet(s)

		p.Debug = s.activeSolutionWord.String()

		fData := TemplateDataForm{}.New(s.language, p, s.PastWords(), s.activeSolutionWord.HasDublicateLetters())
		fData.IsSolved = p.IsSolved()
		fData.IsLoose = p.IsLoose()

		err = t.ExecuteTemplate(w, "lettr-form", fData)
		if err != nil {
			log.Printf("error t.ExecuteTemplate '/lettr' route: %s", err)
		}
	})

	mux.HandleFunc("POST /lettr", func(w http.ResponseWriter, r *http.Request) {
		s := handleSession(w, r, &sessions, wordDb)

		// b, err := io.ReadAll(r.Body)
		// if err != nil {
		// 	// log.Fatalln(err)
		// 	log.Printf("error: %s", err)
		// }
		// log.Printf("word: %s\nbody:\n%s", s.activeWord, b)

		err := r.ParseForm()
		if err != nil {
			log.Printf("error: %s", err)

			w.WriteHeader(422)
			err = t.ExecuteTemplate(w, "oob-messages", TemplateDataMessages{
				ErrMsgs: []Message{"cannot parse form data"},
			})
			if err != nil {
				log.Printf("error t.ExecuteTemplate 'oob-messages': %s", err)
			}
			return
		}

		p := s.lastEvaluatedAttempt
		p.Debug = s.activeSolutionWord.String()

		if p.IsSolved() || p.IsLoose() {
			w.WriteHeader(204)
			return
		}

		if s.lastEvaluatedAttempt.ActiveRow() != countFilledFormRows(r.PostForm)-1 {
			w.WriteHeader(422)
			err = t.ExecuteTemplate(w, "oob-messages", TemplateDataMessages{
				ErrMsgs: []Message{"faked rows"},
			})
			if err != nil {
				log.Printf("error t.ExecuteTemplate 'oob-messages': %s", err)
			}
			return
		}

		p, err = parseForm(p, r.PostForm, s.activeSolutionWord, s.language, wordDb)
		if err == ErrNotInWordList {
			w.WriteHeader(422)
			err = t.ExecuteTemplate(w, "oob-messages", TemplateDataMessages{
				ErrMsgs: []Message{"word not in word list"},
			})
			if err != nil {
				log.Printf("error t.ExecuteTemplate 'oob-messages': %s", err)
			}
			return
		}

		s.lastEvaluatedAttempt = p
		sessions.updateOrSet(s)

		fData := TemplateDataForm{}.New(s.language, p, s.PastWords(), s.activeSolutionWord.HasDublicateLetters())
		fData.IsSolved = p.IsSolved()
		fData.IsLoose = p.IsLoose()

		err = t.ExecuteTemplate(w, "lettr-form", fData)
		if err != nil {
			log.Printf("error t.ExecuteTemplate '/lettr' route: %s", err)
		}
	})

	mux.HandleFunc("POST /new", func(w http.ResponseWriter, r *http.Request) {
		s := handleSession(w, r, &sessions, wordDb)

		// handle lang switch
		l := s.language
		maybeLang := r.FormValue("lang")
		if maybeLang != "" {
			l, _ = NewLang(maybeLang)
			s.language = l

			type TemplateDataLanguge struct {
				Language language
			}
			tData := TemplateDataLanguge{Language: l}

			err := t.ExecuteTemplate(w, "oob-lang-switch", tData)
			if err != nil {
				log.Printf("error t.ExecuteTemplate '/new' route: %s", err)
			}
		}

		p := puzzle.Puzzle{}

		s.lastEvaluatedAttempt = p
		s.AddPastWord(s.activeSolutionWord)
		s.activeSolutionWord = wordDb.RandomPickWithFallback(l, s.pastWords, 0)
		sessions.updateOrSet(s)

		p.Debug = s.activeSolutionWord.String()

		fData := TemplateDataForm{}.New(s.language, p, s.PastWords(), s.activeSolutionWord.HasDublicateLetters())
		fData.IsSolved = p.IsSolved()
		fData.IsLoose = p.IsLoose()

		// w.Header().Add("HX-Refresh", "true")
		err := t.ExecuteTemplate(w, "lettr-form", fData)
		if err != nil {
			log.Printf("error t.ExecuteTemplate '/new' route: %s", err)
		}
	})

	mux.HandleFunc("POST /help", func(w http.ResponseWriter, r *http.Request) {
		s := handleSession(w, r, &sessions, wordDb)

		p := s.lastEvaluatedAttempt

		sessions.updateOrSet(s)

		p.Debug = s.activeSolutionWord.String()

		fData := TemplateDataForm{}.New(s.language, p, s.PastWords(), s.activeSolutionWord.HasDublicateLetters())
		fData.IsSolved = p.IsSolved()
		fData.IsLoose = p.IsLoose()

		err := t.ExecuteTemplate(w, "help", fData)
		if err != nil {
			log.Printf("error t.ExecuteTemplate '/help' route: %s", err)
		}
	})

	mux.HandleFunc("GET /suggest", func(w http.ResponseWriter, r *http.Request) {
		err := t.ExecuteTemplate(w, "suggest", TemplateDataSuggest{})
		if err != nil {
			log.Printf("error t.ExecuteTemplate '/suggest' route: %s", err)
		}
	})

	mux.HandleFunc("POST /suggest", func(w http.ResponseWriter, r *http.Request) {
		var tdm TemplateDataMessages

		err := r.ParseForm()
		if err != nil {
			log.Printf("error: %s", err)

			w.WriteHeader(422)
			err = t.ExecuteTemplate(w, "oob-messages", TemplateDataMessages{
				ErrMsgs: []Message{"can not parse form data"},
			})
			if err != nil {
				log.Printf("error t.ExecuteTemplate 'oob-messages': %s", err)
			}
			return
		}

		form := r.PostForm
		tds := TemplateDataSuggest{
			Word:     form["word"][0],
			Message:  form["message"][0],
			Language: form["language-pick"][0],
			Action:   form["suggest-action"][0],
		}

		err = tds.validate()
		if err != nil {
			w.WriteHeader(422)
			w.Header().Add("HX-Reswap", "none")

			tdm = TemplateDataMessages{
				ErrMsgs: []Message{Message(err.Error())},
			}

			err = t.ExecuteTemplate(w, "oob-messages", tdm)
			if err != nil {
				log.Printf("error t.ExecuteTemplate 'oob-messages' route: %s", err)
			}

			return
		}

		err = github.CreateWordSuggestionIssue(
			context.Background(), envCfg.githubToken, tds.Word, tds.Language, tds.Action, tds.Message,
		)
		if err != nil {
			w.WriteHeader(422)
			w.Header().Add("HX-Reswap", "none")

			tdm = TemplateDataMessages{
				ErrMsgs: []Message{"Could not send suggestion."},
			}

			err = t.ExecuteTemplate(w, "oob-messages", tdm)
			if err != nil {
				log.Printf("error t.ExecuteTemplate 'oob-messages' route: %s", err)
			}

			return
		}

		tdm = TemplateDataMessages{
			SuccessMsgs: []Message{"Suggestion send, thank you!"},
		}
		err = t.ExecuteTemplate(w, "oob-messages", tdm)
		if err != nil {
			log.Printf("error t.ExecuteTemplate 'oob-messages' route: %s", err)
		}

		err = t.ExecuteTemplate(w, "suggest", TemplateDataSuggest{})
		if err != nil {
			log.Printf("error t.ExecuteTemplate '/suggest' route: %s", err)
		}
	})

	counter := counterState{count: 0}
	mux.HandleFunc("POST /counter", func(w http.ResponseWriter, req *http.Request) {
		// handleSession(w, req, &sessions)
		counter.mu.Lock()
		counter.count++
		defer counter.mu.Unlock()

		b, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("Method: %s\nbody:\n%s", req.Method, b)

		io.WriteString(w, fmt.Sprintf("<span>%d</span>", counter.count))

	})

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
		panic("GITHUB_TOKEN not provided")
	}

	return env{port: port, githubToken: gt}
}

func handleSession(w http.ResponseWriter, req *http.Request, sessions *sessions, wdb wordDatabase) session {
	var err error
	var sess session

	cookie, err := req.Cookie(SESSION_COOKIE_NAME)
	if err != nil {
		return newSession(w, sessions, wdb)
	}

	if cookie == nil {
		return newSession(w, sessions, wdb)
	}

	sid := cookie.Value
	i := slices.IndexFunc(*sessions, func(s session) bool {
		return s.id == sid
	})
	if i == -1 {
		return newSession(w, sessions, wdb)
	}

	sess = (*sessions)[i]

	c := constructCookie(sess)
	http.SetCookie(w, &c)

	sess.expiresAt = generateSessionLifetime()
	(*sessions)[i] = sess

	return sess
}

func newSession(w http.ResponseWriter, sessions *sessions, wdb wordDatabase) session {
	sess := generateSession(LANG_EN, wdb)
	*sessions = append(*sessions, sess)
	c := constructCookie(sess)
	http.SetCookie(w, &c)

	return sess
}

func constructCookie(s session) http.Cookie {
	return http.Cookie{
		Name:     SESSION_COOKIE_NAME,
		Value:    s.id,
		Path:     "/",
		MaxAge:   s.maxAgeSeconds,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func generateSession(lang language, wdb wordDatabase) session { //todo: pass it by ref not by copy?
	id := uuid.NewString()
	expiresAt := generateSessionLifetime()
	activeWord, err := wdb.RandomPick(lang, []puzzle.Word{}, 0)
	if err != nil {
		log.Printf("pick random word failed: %s", err)

		activeWord = puzzle.Word{'R', 'O', 'A', 'T', 'E'}.ToLower()
	}

	return session{id, expiresAt, SESSION_MAX_AGE_IN_SECONDS, lang, activeWord, puzzle.Puzzle{}, []puzzle.Word{}}
}

func generateSessionLifetime() time.Time {
	return time.Now().Add(SESSION_MAX_AGE_IN_SECONDS * time.Second) // todo: 24 hour, sec now only for testing
}

func countFilledFormRows(postPuzzleForm url.Values) uint8 {
	isfilled := func(row []string) bool {
		emptyButWithLen := make([]string, len(row)) // we need empty slice but with right elem length
		return !(slices.Compare(row, emptyButWithLen) == 0)
	}

	var count uint8 = 0
	l := len(postPuzzleForm)
	for i := 0; i < l; i++ {
		guessedWord, ok := postPuzzleForm[fmt.Sprintf("r%d", i)]
		if ok && isfilled(guessedWord) {
			count++
		}
	}

	return count
}

func parseForm(p puzzle.Puzzle, form url.Values, solutionWord puzzle.Word, l language, wdb wordDatabase) (puzzle.Puzzle, error) {
	for ri := range p.Guesses {
		maybeGuessedWord, ok := form[fmt.Sprintf("r%d", ri)]
		if !ok {
			continue
		}

		guessedWord, err := sliceToWord(maybeGuessedWord)
		if err != nil {
			return p, fmt.Errorf("parseForm could not create guessedWord from form input: %s", err.Error())
		}

		if !wdb.Exists(l, guessedWord) {
			return p, ErrNotInWordList
		}

		wg := evaluateGuessedWord(guessedWord, solutionWord)

		p.Guesses[ri] = wg
	}

	return p, nil
}

func sliceToWord(maybeGuessedWord []string) (puzzle.Word, error) {
	w := puzzle.Word{}

	if len(maybeGuessedWord) != len(w) {
		return puzzle.Word{}, fmt.Errorf("sliceToWord: provided slice does not match word length")
	}

	for i, l := range maybeGuessedWord {
		w[i], _ = utf8.DecodeRuneInString(strings.ToLower(l))
		if w[i] == 65533 {
			w[i] = 0
		}
	}

	return w, nil
}

func evaluateGuessedWord(guessedWord puzzle.Word, solutionWord puzzle.Word) puzzle.WordGuess {
	solutionWord = solutionWord.ToLower()
	guessedLetterCountMap := make(map[rune]int)

	resultWordGuess := puzzle.WordGuess{}

	// initilize
	for i, gr := range guessedWord {
		resultWordGuess[i].Letter = gr
		resultWordGuess[i].Match = puzzle.MatchNone
	}

	// mark exact matches
	for i, gr := range guessedWord {
		exact := solutionWord[i] == gr

		if exact {
			guessedLetterCountMap[gr]++
			resultWordGuess[i].Match = puzzle.MatchExact
		}
	}

	// mark some/vague matches
	for i, gr := range guessedWord {
		if resultWordGuess[i].Match == puzzle.MatchExact {
			continue
		}

		some := solutionWord.Contains(gr)

		if !(resultWordGuess[i].Match == puzzle.MatchVague) || some {
			guessedLetterCountMap[gr]++
		}

		s := some && (guessedLetterCountMap[gr] <= solutionWord.Count(gr))
		if s {
			resultWordGuess[i].Match = puzzle.MatchVague
		}
	}

	return resultWordGuess
}
