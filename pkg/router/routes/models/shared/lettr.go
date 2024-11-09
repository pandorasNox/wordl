package shared

import (
	"github.com/pandorasNox/lettr/pkg/language"
	"github.com/pandorasNox/lettr/pkg/puzzle"
)

type TemplateDataLettr struct {
	Data        puzzle.Puzzle
	IsSolved    bool
	IsLoose     bool
	Language    language.Language
	Revision    string
	FaviconPath string
	Keyboard    Keyboard
	PastWords   []puzzle.Word
	ImprintUrl  string
}

func (fd TemplateDataLettr) New(l language.Language, p puzzle.Puzzle, letterHints []rune, pastWords []puzzle.Word, imprintUrl string, revision string, faviconPath string) TemplateDataLettr {
	kb := Keyboard{}
	kb.Init(l, p.LetterGuesses(), letterHints)

	return TemplateDataLettr{
		Data:        p,
		Language:    l,
		Revision:    revision,
		FaviconPath: faviconPath,
		Keyboard:    kb,
		PastWords:   pastWords,
		ImprintUrl:  imprintUrl,
	}
}
