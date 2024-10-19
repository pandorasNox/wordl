package shared

import (
	"time"

	"github.com/pandorasNox/lettr/pkg/language"
	"github.com/pandorasNox/lettr/pkg/puzzle"
)

type TemplateDataLettr struct {
	Data                        puzzle.Puzzle
	Errors                      map[string]string
	IsSolved                    bool
	IsLoose                     bool
	JSCachePurgeTimestamp       int64
	Language                    language.Language
	Revision                    string
	FaviconPath                 string
	Keyboard                    Keyboard
	PastWords                   []puzzle.Word
	SolutionHasDublicateLetters bool
	ImprintUrl                  string
}

func (fd TemplateDataLettr) New(l language.Language, p puzzle.Puzzle, pastWords []puzzle.Word, solutionHasDublicateLetters bool, imprintUrl string, revision string, faviconPath string) TemplateDataLettr {
	kb := Keyboard{}
	kb.Init(l, p.LetterGuesses())

	return TemplateDataLettr{
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
