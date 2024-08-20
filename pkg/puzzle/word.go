package puzzle

import (
	"fmt"
	"unicode"
)

type Word [5]rune

func (w Word) String() string {
	out := ""
	for _, v := range w {
		out += string(v)
	}

	return out
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
