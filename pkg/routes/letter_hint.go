package routes

import (
	"log"
	"math/rand"
	"net/http"
	"slices"
	"time"

	"github.com/pandorasNox/lettr/pkg/notification"
	"github.com/pandorasNox/lettr/pkg/puzzle"
	"github.com/pandorasNox/lettr/pkg/routes/models"
	"github.com/pandorasNox/lettr/pkg/session"
)

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

func LetterHint(sessions *session.Sessions, wdb puzzle.WordDatabase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		notifier := notification.NewNotifier()
		sess := session.HandleSession(w, r, sessions, wdb)
		gameState := sess.GameState()

		solutionWord := gameState.ActiveSolutionWord()
		lg := gameState.LastEvaluatedAttempt().LetterGuesses()

		matchedLetter := slices.DeleteFunc(lg, func(l puzzle.LetterGuess) bool {
			return l.Match.Is(puzzle.MatchNone)
		})
		matchedRunes := Map(matchedLetter, func(l puzzle.LetterGuess) rune {
			return l.Letter
		})

		notFoundLetters := slices.DeleteFunc(solutionWord.ToSlice(), func(l rune) bool {
			return slices.Contains(matchedRunes, l)
		})

		lhs := gameState.LetterHints()
		hintOptions := slices.DeleteFunc(notFoundLetters, func(l rune) bool {
			return slices.Contains(lhs, l)
		})

		randSrc := rand.NewSource(time.Now().UnixNano())
		pick := PickRandomRune(hintOptions, randSrc)
		if pick == rune(0) {
			notifier.AddInfo("No more hints to provide")
			err := routesTemplates.ExecuteTemplate(w, "oob-messages", notifier.ToTemplate())
			if err != nil {
				log.Printf("error t.ExecuteTemplate 'oob-messages': %s", err)
			}
			return
		}

		gameState.AddLetterHint(pick)
		sessions.UpdateOrSet(sess)

		err := routesTemplates.ExecuteTemplate(w, "single-letter-hint", models.TemplateDataLetterHint(pick))
		if err != nil {
			log.Printf("error t.ExecuteTemplate 'single-letter-hint': %s", err)
		}
	}
}
