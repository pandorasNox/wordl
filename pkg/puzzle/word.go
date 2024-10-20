package puzzle

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Word [5]rune

func (w Word) String() string {
	out := ""
	for _, v := range w {
		out += string(v)
	}

	return out
}

func (w Word) ToSlice() []rune {
	o := []rune{}

	for _, v := range w {
		o = append(o, v)
	}

	return o
}

func (w Word) Contains(letter rune) bool {
	found := false
	for _, v := range w {
		if v == letter {
			found = true
			break
		}
	}

	return found
}

func (w Word) Count(letter rune) int {
	count := 0
	for _, v := range w {
		if v == letter {
			count++
		}
	}

	return count
}

func (w Word) IsEqual(compare Word) bool {
	for i, v := range w {
		if v != compare[i] {
			return false
		}
	}

	return true
}

func (w Word) HasDublicateLetters() bool {
	for _, l := range w {
		if w.Count(l) >= 2 {
			return true
		}
	}

	return false
}

func (w Word) ToLower() Word {
	for i, v := range w {
		w[i] = unicode.ToLower(v)
	}

	return w
}

func toWord(wo string) (Word, error) {
	out := Word{}

	length := 0
	for i, l := range wo {
		length++
		if length > len(out) {
			return Word{}, fmt.Errorf("string does not match allowed word length: length=%d, expectedLength=%d", length, len(out))
		}

		out[i] = l
	}

	if length < len(out) {
		return Word{}, fmt.Errorf("string is to short: length=%d, expectedLength=%d", length, len(out))
	}

	return out, nil
}

func SliceToWord(maybeGuessedWord []string) (Word, error) {
	w := Word{}

	if len(maybeGuessedWord) != len(w) {
		return Word{}, fmt.Errorf("sliceToWord: provided slice does not match word length")
	}

	for i, l := range maybeGuessedWord {
		w[i], _ = utf8.DecodeRuneInString(strings.ToLower(l))
		if w[i] == 65533 {
			w[i] = 0
		}
	}

	return w, nil
}
