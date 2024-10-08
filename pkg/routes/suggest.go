package routes

import (
	"context"
	"errors"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"slices"

	"github.com/microcosm-cc/bluemonday"
	"github.com/pandorasNox/lettr/pkg/github"
	"github.com/pandorasNox/lettr/pkg/notification"
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

func (tds TemplateDataSuggest) validate() error {
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

func GetSuggest(t *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := t.ExecuteTemplate(w, "suggest", TemplateDataSuggest{})
		if err != nil {
			log.Printf("error t.ExecuteTemplate '/suggest' route: %s", err)
		}
	}
}

func PostSuggest(t *template.Template, githubToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		notifier := notification.NewNotifier()

		err := r.ParseForm()
		if err != nil {
			log.Printf("error: %s", err)

			w.WriteHeader(422)
			notifier.AddError("can not parse form data")
			err = t.ExecuteTemplate(w, "oob-messages", notifier.ToTemplate())
			if err != nil {
				log.Printf("error t.ExecuteTemplate 'oob-messages': %s", err)
			}
			return
		}

		form := r.PostForm
		tds := TemplateDataSuggest{
			Word:     form["word"][0],
			Message:  form["message"][0],
			Language: form["language-pick"][0],
			Action:   form["suggest-action"][0],
		}

		err = tds.validate()
		if err != nil {
			w.WriteHeader(422)
			w.Header().Add("HX-Reswap", "none")

			notifier.AddError(err.Error())
			err = t.ExecuteTemplate(w, "oob-messages", notifier.ToTemplate())
			if err != nil {
				log.Printf("error t.ExecuteTemplate 'oob-messages' route: %s", err)
			}

			return
		}

		err = github.CreateWordSuggestionIssue(
			context.Background(), githubToken, tds.Word, tds.Language, tds.Action, tds.Message,
		)
		if err != nil {
			w.WriteHeader(422)
			w.Header().Add("HX-Reswap", "none")

			notifier.AddError("Could not send suggestion.")
			err = t.ExecuteTemplate(w, "oob-messages", notifier.ToTemplate())
			if err != nil {
				log.Printf("error t.ExecuteTemplate 'oob-messages' route: %s", err)
			}

			return
		}

		notifier.AddSuccess("Suggestion send, thank you!")
		err = t.ExecuteTemplate(w, "oob-messages", notifier.ToTemplate())
		if err != nil {
			log.Printf("error t.ExecuteTemplate 'oob-messages' route: %s", err)
		}

		err = t.ExecuteTemplate(w, "suggest", TemplateDataSuggest{})
		if err != nil {
			log.Printf("error t.ExecuteTemplate '/suggest' route: %s", err)
		}
	}
}
