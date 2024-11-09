package shared

import (
	"slices"
	"unicode"

	"github.com/pandorasNox/lettr/pkg/language"
	"github.com/pandorasNox/lettr/pkg/puzzle"
)

type keyboardKey struct {
	Key    string
	IsUsed bool
	Match  puzzle.Match
	IsHint bool
}

type Keyboard struct {
	KeyGrid [][]keyboardKey
}

func (k *Keyboard) Init(l language.Language, lgs []puzzle.LetterGuess, letterHints []rune) {
	k.KeyGrid = [][]keyboardKey{
		{{"Q", false, puzzle.MatchNone, false}, {"W", false, puzzle.MatchNone, false}, {"E", false, puzzle.MatchNone, false}, {"R", false, puzzle.MatchNone, false}, {"T", false, puzzle.MatchNone, false}, {"Y", false, puzzle.MatchNone, false}, {"U", false, puzzle.MatchNone, false}, {"I", false, puzzle.MatchNone, false}, {"O", false, puzzle.MatchNone, false}, {"P", false, puzzle.MatchNone, false}, {"Delete", false, puzzle.MatchNone, false}},
		{{"A", false, puzzle.MatchNone, false}, {"S", false, puzzle.MatchNone, false}, {"D", false, puzzle.MatchNone, false}, {"F", false, puzzle.MatchNone, false}, {"G", false, puzzle.MatchNone, false}, {"H", false, puzzle.MatchNone, false}, {"J", false, puzzle.MatchNone, false}, {"K", false, puzzle.MatchNone, false}, {"L", false, puzzle.MatchNone, false}, {"Enter", false, puzzle.MatchNone, false}},
		{{"Z", false, puzzle.MatchNone, false}, {"X", false, puzzle.MatchNone, false}, {"C", false, puzzle.MatchNone, false}, {"V", false, puzzle.MatchNone, false}, {"B", false, puzzle.MatchNone, false}, {"N", false, puzzle.MatchNone, false}, {"M", false, puzzle.MatchNone, false}},
	}

	for ri, keyboardRow := range k.KeyGrid {
		for ki, currentKey := range keyboardRow {
			// ensure length is one (for to rune conversion, skipping keys like "Enter" or "Delete")
			if len(currentKey.Key) > 1 {
				continue
			}

			KeyRune := unicode.ToLower(firstRune(currentKey.Key))

			isExactMatch := slices.ContainsFunc(lgs, func(lg puzzle.LetterGuess) bool {
				return KeyRune == lg.Letter && lg.Match == puzzle.MatchExact
			})
			if isExactMatch {
				k.KeyGrid[ri][ki].IsUsed = true
				k.KeyGrid[ri][ki].Match = puzzle.MatchExact
				continue
			}

			isVagueMatch := slices.ContainsFunc(lgs, func(lg puzzle.LetterGuess) bool {
				return KeyRune == lg.Letter && lg.Match == puzzle.MatchVague
			})
			if isVagueMatch {
				k.KeyGrid[ri][ki].IsUsed = true
				k.KeyGrid[ri][ki].Match = puzzle.MatchVague
				continue
			}

			isHint := slices.Contains(letterHints, KeyRune)
			k.KeyGrid[ri][ki].IsHint = isHint

			isUsed := slices.ContainsFunc(lgs, func(lg puzzle.LetterGuess) bool {
				return KeyRune == lg.Letter && lg.Match == puzzle.MatchNone
			})
			if isUsed {
				k.KeyGrid[ri][ki].IsUsed = true
				k.KeyGrid[ri][ki].Match = puzzle.MatchNone
				continue
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
