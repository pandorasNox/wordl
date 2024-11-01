package models

import "github.com/pandorasNox/lettr/pkg/puzzle"

type TemplateDataHelpPage struct {
	SolutionWord                string
	SolutionHasDublicateLetters bool
	LetterHints                 []rune
	PastWords                   []puzzle.Word
}
