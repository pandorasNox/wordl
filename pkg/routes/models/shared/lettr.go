package shared

import (
	"time"

	"github.com/pandorasNox/lettr/pkg/language"
	"github.com/pandorasNox/lettr/pkg/puzzle"
)

type TemplateDataLettr struct {
	Data                  puzzle.Puzzle
	IsSolved              bool
	IsLoose               bool
	JSCachePurgeTimestamp int64
	Language              language.Language
	Revision              string
	FaviconPath           string
	Keyboard              Keyboard
	PastWords             []puzzle.Word
	ImprintUrl            string
}

func (fd TemplateDataLettr) New(l language.Language, p puzzle.Puzzle, pastWords []puzzle.Word, imprintUrl string, revision string, faviconPath string) TemplateDataLettr {
	kb := Keyboard{}
	kb.Init(l, p.LetterGuesses())

	return TemplateDataLettr{
		Data:                  p,
		JSCachePurgeTimestamp: time.Now().Unix(),
		Language:              l,
		Revision:              revision,
		FaviconPath:           faviconPath,
		Keyboard:              kb,
		PastWords:             pastWords,
		ImprintUrl:            imprintUrl,
	}
}
