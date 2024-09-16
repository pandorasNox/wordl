package routes

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"slices"
	"time"

	"github.com/pandorasNox/lettr/pkg/notification"
	"github.com/pandorasNox/lettr/pkg/puzzle"
	"github.com/pandorasNox/lettr/pkg/session"
)

type TemplateDataLetterHint struct{}

// FilterFunc is ...
func FilterFunc[S ~[]E, E any](s S, fnShouldKeep func(E) bool) S {
	o := S{}

	for _, v := range s {
		if fnShouldKeep(v) {
			o = append(o, v)
		}
	}

	return o
}

func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

func PickRandomRune(runeList []rune, randSrc rand.Source) rune {
	if len(runeList) == 0 {
		return rune(0)
	}
	if len(runeList) == 1 {
		return runeList[0]
	}

	randgenerator := rand.New(randSrc)
	randIndex := randgenerator.Intn(len(runeList))

	return runeList[randIndex]
}

func LetterHint(t *template.Template, sessions *session.Sessions, wdb puzzle.WordDatabase) http.HandlerFunc {
	notifier := notification.NewNotifier()

	return func(w http.ResponseWriter, r *http.Request) {
		sess := session.HandleSession(w, r, sessions, wdb)

		solutionWord := sess.ActiveSolutionWord()
		lg := sess.LastEvaluatedAttempt().LetterGuesses()

		matchedLetter := FilterFunc(lg, func(l puzzle.LetterGuess) bool {
			return l.Match.Is(puzzle.MatchExact) || l.Match.Is(puzzle.MatchVague)
		})
		matchedRunes := Map(matchedLetter, func(l puzzle.LetterGuess) rune {
			return l.Letter
		})

		notFoundLetters := FilterFunc(solutionWord.ToSlice(), func(l rune) bool {
			return !slices.Contains(matchedRunes, l)
		})

		lhs := sess.LetterHints()
		hintOptions := FilterFunc(notFoundLetters, func(l rune) bool {
			return !slices.Contains(lhs, l)
		})

		randSrc := rand.NewSource(time.Now().UnixNano())
		pick := PickRandomRune(hintOptions, randSrc)
		if pick == rune(0) {
			notifier.AddInfo("No more hints to provide")
			err := t.ExecuteTemplate(w, "oob-messages", notifier.ToTemplate())
			if err != nil {
				log.Printf("error t.ExecuteTemplate 'oob-messages': %s", err)
			}
			return
		}

		sess.AddLetterHint(pick)
		sessions.UpdateOrSet(sess)

		w.Write([]byte{byte(pick)})
	}
}
