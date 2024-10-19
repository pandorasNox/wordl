package language

import "fmt"

type Language string

const (
	LANG_EN Language = "en"
	LANG_DE Language = "de"
)

func NewLang(maybeLang string) (Language, error) {
	switch Language(maybeLang) {
	case LANG_EN, LANG_DE:
		return Language(maybeLang), nil
	default:
		return LANG_EN, fmt.Errorf("couldn't create new language from given value: '%s'", maybeLang)
	}
}
