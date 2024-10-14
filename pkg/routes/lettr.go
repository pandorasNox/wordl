package routes

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/pandorasNox/lettr/pkg/language"
	"github.com/pandorasNox/lettr/pkg/notification"
	"github.com/pandorasNox/lettr/pkg/puzzle"
	"github.com/pandorasNox/lettr/pkg/session"
)

var ErrNotInWordList = errors.New("not in wordlist")

type TemplateDataLettrForm struct {
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

func (fd TemplateDataLettrForm) New(
	l language.Language,
	p puzzle.Puzzle,
	pastWords []puzzle.Word,
	solutionHasDublicateLetters bool,
	imprintUrl string,
	revision string,
	faviconPath string,
) TemplateDataLettrForm {
	kb := keyboard{}
	kb.Init(l, p.LetterGuesses())

	return TemplateDataLettrForm{
		Data:                        p,
		Errors:                      make(map[string]string),
		JSCachePurgeTimestamp:       time.Now().Unix(),
		Language:                    l,
		Revision:                    revision,
		FaviconPath:                 faviconPath,
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

func GetLettr(t *template.Template, sessions *session.Sessions, wdb puzzle.WordDatabase, imprintUrl string, revision string, faviconPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := session.HandleSession(w, r, sessions, wdb)
		sessions.UpdateOrSet(s)

		p := s.GameState().LastEvaluatedAttempt()

		fData := TemplateDataLettrForm{}.New(
			s.Language(),
			p,
			s.PastWords(),
			s.GameState().ActiveSolutionWord().HasDublicateLetters(),
			imprintUrl,
			revision,
			faviconPath,
		)
		fData.IsSolved = p.IsSolved()
		fData.IsLoose = p.IsLoose()

		err := t.ExecuteTemplate(w, "lettr-form", fData)
		if err != nil {
			log.Printf("error t.ExecuteTemplate '/lettr' route: %s", err)
		}
	}
}

func PostLettr(t *template.Template, sessions *session.Sessions, wdb puzzle.WordDatabase, imprintUrl string, revision string, faviconPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := session.HandleSession(w, r, sessions, wdb)
		notifier := notification.NewNotifier()

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
			notifier.AddError("cannot parse form data")
			err = t.ExecuteTemplate(w, "oob-messages", notifier.ToTemplate())
			if err != nil {
				log.Printf("error t.ExecuteTemplate 'oob-messages': %s", err)
			}
			return
		}

		g := s.GameState()
		p := g.LastEvaluatedAttempt()

		if p.IsSolved() || p.IsLoose() {
			w.WriteHeader(204)
			return
		}

		if p.ActiveRow() != countFilledFormRows(r.PostForm)-1 {
			w.WriteHeader(422)
			notifier.AddError("faked rows")
			err = t.ExecuteTemplate(w, "oob-messages", notifier.ToTemplate())
			if err != nil {
				log.Printf("error t.ExecuteTemplate 'oob-messages': %s", err)
			}
			return
		}

		p, err = parseForm(p, r.PostForm, g.ActiveSolutionWord(), s.Language(), wdb)
		if err == ErrNotInWordList {
			w.WriteHeader(422)
			notifier.AddError("word not in word list")
			err = t.ExecuteTemplate(w, "oob-messages", notifier.ToTemplate())
			if err != nil {
				log.Printf("error t.ExecuteTemplate 'oob-messages': %s", err)
			}
			return
		}

		g.SetLastEvaluatedAttempt(p)
		s.SetGameState(*g) //todo move gamestate from pointer to copy
		sessions.UpdateOrSet(s)

		fData := TemplateDataLettrForm{}.New(s.Language(), p, s.PastWords(), g.ActiveSolutionWord().HasDublicateLetters(), imprintUrl, revision, faviconPath)
		fData.IsSolved = p.IsSolved()
		fData.IsLoose = p.IsLoose()

		err = t.ExecuteTemplate(w, "lettr-form", fData)
		if err != nil {
			log.Printf("error t.ExecuteTemplate '/lettr' route: %s", err)
		}
	}
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

func parseForm(p puzzle.Puzzle, form url.Values, solutionWord puzzle.Word, l language.Language, wdb puzzle.WordDatabase) (puzzle.Puzzle, error) {
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
