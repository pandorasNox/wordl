package routes

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/pandorasNox/lettr/pkg/language"
	"github.com/pandorasNox/lettr/pkg/puzzle"
)

func Test_parseForm(t *testing.T) {
	type args struct {
		p            puzzle.Puzzle
		form         url.Values
		solutionWord puzzle.Word
		language     language.Language
		wdb          puzzle.WordDatabase
	}
	tests := []struct {
		name    string
		args    args
		want    puzzle.Puzzle
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "no hits, neither same or exact",
			// args: args{puzzle{}, url.Values{}, word{'M', 'I', 'S', 'S', 'S'}},
			args: args{
				p:            puzzle.Puzzle{},
				form:         url.Values{"r0": make([]string, 5)},
				solutionWord: puzzle.Word{'M', 'I', 'S', 'S', 'S'},
				language:     language.LANG_EN,
				wdb: puzzle.WordDatabase{Db: map[language.Language]map[puzzle.WordCollection]map[puzzle.Word]bool{
					language.LANG_EN: {
						puzzle.WC_COMMON: {
							puzzle.Word{'m', 'i', 's', 's', 's'}: true,
							puzzle.Word{0, 0, 0, 0, 0}:           true, // equals make([]string, 5)
						},
						puzzle.WC_ALL: {
							puzzle.Word{'m', 'i', 's', 's', 's'}: true,
							puzzle.Word{0, 0, 0, 0, 0}:           true, // equals make([]string, 5)
						},
					},
				}},
			},
			want: puzzle.Puzzle{
				Guesses: [6]puzzle.WordGuess{
					{
						puzzle.LetterGuess{Match: puzzle.MatchNone},
						puzzle.LetterGuess{Match: puzzle.MatchNone},
						puzzle.LetterGuess{Match: puzzle.MatchNone},
						puzzle.LetterGuess{Match: puzzle.MatchNone},
						puzzle.LetterGuess{Match: puzzle.MatchNone},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "full exact match",
			args: args{
				p:            puzzle.Puzzle{},
				form:         url.Values{"r0": []string{"M", "A", "T", "C", "H"}},
				solutionWord: puzzle.Word{'M', 'A', 'T', 'C', 'H'},
				language:     language.LANG_EN,
				wdb: puzzle.WordDatabase{Db: map[language.Language]map[puzzle.WordCollection]map[puzzle.Word]bool{
					language.LANG_EN: {
						puzzle.WC_COMMON: {
							puzzle.Word{'m', 'a', 't', 'c', 'h'}: true,
						},
						puzzle.WC_ALL: {
							puzzle.Word{'m', 'a', 't', 'c', 'h'}: true,
						},
					},
				}},
			},
			want: puzzle.Puzzle{Guesses: [6]puzzle.WordGuess{
				{
					{Letter: 'm', Match: puzzle.MatchExact},
					{Letter: 'a', Match: puzzle.MatchExact},
					{Letter: 't', Match: puzzle.MatchExact},
					{Letter: 'c', Match: puzzle.MatchExact},
					{Letter: 'h', Match: puzzle.MatchExact},
				},
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := parseForm(tt.args.p, tt.args.form, tt.args.solutionWord, tt.args.language, tt.args.wdb); !reflect.DeepEqual(got, tt.want) || (err != nil) != tt.wantErr {
				t.Errorf("parseForm() = %v, %v; want %v, %v", got, err != nil, tt.want, tt.wantErr)
			}
		})
	}
}
