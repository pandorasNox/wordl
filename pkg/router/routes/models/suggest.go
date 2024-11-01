package models

import (
	"errors"
	"regexp"
	"slices"

	"github.com/microcosm-cc/bluemonday"
)

type TemplateDataSuggest struct {
	Word     string
	Message  string
	Language string
	Action   string
}

var RegexpAllowedWordCharacters = regexp.MustCompile(`^[A-Za-zöäüÖÄÜß]{5}$`)

var ErrFailedWordValidation = errors.New("validation failed: word is either to long, to short or contains forbidden characters")
var ErrFailedMessageValidation = errors.New("validation failed: message contains invalid data")
var ErrFailedActionValidation = errors.New("validation failed: action invalid")
var ErrFailedLanguageValidation = errors.New("validation failed: language invalid")

func (tds TemplateDataSuggest) Validate() error {
	if !RegexpAllowedWordCharacters.Match([]byte(tds.Word)) {
		return ErrFailedWordValidation
	}

	p := bluemonday.UGCPolicy()
	sm := p.Sanitize(tds.Message)
	if sm != tds.Message {
		return ErrFailedMessageValidation
	}

	if !slices.Contains([]string{"add", "remove"}, tds.Action) {
		return ErrFailedActionValidation
	}

	if tds.Language != "german" && tds.Language != "english" {
		return ErrFailedLanguageValidation
	}

	return nil
}
