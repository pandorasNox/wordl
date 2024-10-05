package main

import (
	iofs "io/fs"
	"net/url"
	"reflect"
	"slices"
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
			want: puzzle.Puzzle{Debug: "", Guesses: [6]puzzle.WordGuess{
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

func Test_evaluateGuessedWord(t *testing.T) {
	type args struct {
		guessedWord  puzzle.Word
		solutionWord puzzle.Word
	}
	tests := []struct {
		name string
		args args
		want puzzle.WordGuess
	}{
		// test cases
		{
			name: "no hits, neither same or exact",
			args: args{
				guessedWord:  puzzle.Word{},
				solutionWord: puzzle.Word{'M', 'I', 'S', 'S', 'S'},
			},
			want: puzzle.WordGuess{
				{Match: puzzle.MatchNone},
				{Match: puzzle.MatchNone},
				{Match: puzzle.MatchNone},
				{Match: puzzle.MatchNone},
				{Match: puzzle.MatchNone},
			},
		},
		{
			name: "full exact match",
			args: args{
				guessedWord:  puzzle.Word{'m', 'a', 't', 'c', 'h'},
				solutionWord: puzzle.Word{'M', 'A', 'T', 'C', 'H'},
			},
			want: puzzle.WordGuess{
				{Letter: 'm', Match: puzzle.MatchExact},
				{Letter: 'a', Match: puzzle.MatchExact},
				{Letter: 't', Match: puzzle.MatchExact},
				{Letter: 'c', Match: puzzle.MatchExact},
				{Letter: 'h', Match: puzzle.MatchExact},
			},
		},
		{
			name: "partial exact and partial some match",
			args: args{
				guessedWord:  puzzle.Word{'r', 'a', 'u', 'l', 'o'},
				solutionWord: puzzle.Word{'R', 'O', 'A', 'T', 'E'},
			},
			want: puzzle.WordGuess{
				{Letter: 'r', Match: puzzle.MatchExact},
				{Letter: 'a', Match: puzzle.MatchVague},
				{Letter: 'u', Match: puzzle.MatchNone},
				{Letter: 'l', Match: puzzle.MatchNone},
				{Letter: 'o', Match: puzzle.MatchVague},
			},
		},
		{
			name: "guessed word contains duplicats",
			args: args{
				guessedWord:  puzzle.Word{'r', 'o', 't', 'o', 'r'},
				solutionWord: puzzle.Word{'R', 'O', 'A', 'T', 'E'},
			},
			want: puzzle.WordGuess{
				{Letter: 'r', Match: puzzle.MatchExact},
				{Letter: 'o', Match: puzzle.MatchExact},
				{Letter: 't', Match: puzzle.MatchVague},
				{Letter: 'o', Match: puzzle.MatchNone}, // both false bec we already found it or even already guesst the exact match
				{Letter: 'r', Match: puzzle.MatchNone}, // both false bec we already found it or even already guesst the exact match
			},
		},
		{
			name: "guessed word contains duplicats at end",
			args: args{
				guessedWord:  puzzle.Word{'i', 'x', 'i', 'i', 'i'},
				solutionWord: puzzle.Word{'L', 'X', 'I', 'I', 'I'},
			},
			want: puzzle.WordGuess{
				{Letter: 'i', Match: puzzle.MatchNone},
				{Letter: 'x', Match: puzzle.MatchExact},
				{Letter: 'i', Match: puzzle.MatchExact},
				{Letter: 'i', Match: puzzle.MatchExact},
				{Letter: 'i', Match: puzzle.MatchExact},
			},
		},
		{
			name: "guessed word contains duplicats at end fpp",
			args: args{
				guessedWord:  puzzle.Word{'l', 'i', 'i', 'i', 'i'},
				solutionWord: puzzle.Word{'I', 'L', 'X', 'I', 'I'},
			},
			want: puzzle.WordGuess{
				{Letter: 'l', Match: puzzle.MatchVague},
				{Letter: 'i', Match: puzzle.MatchVague},
				{Letter: 'i', Match: puzzle.MatchNone},
				{Letter: 'i', Match: puzzle.MatchExact},
				{Letter: 'i', Match: puzzle.MatchExact},
			},
		},
		// {
		// 	name: "target word contains duplicats / guessed word contains duplicats",
		// 	args: args{
		// 		puzzle.Puzzle{},
		// 		url.Values{"r0c0": []string{"M"}, "r0c1": []string{"A"}, "r0c2": []string{"T"}, "r0c3": []string{"C"}, "r0c4": []string{"H"}},
		// 		word{'M', 'A', 'T', 'C', 'H'},
		// 	},
		// 	want: puzzle.Puzzle{"", puzzle.wordGuess{
		// 		{
		// 			{'r', puzzle.LetterExact},
		// 			{'o', puzzle.LetterExact},
		// 			{'t', puzzle.LetterExact},
		// 			{'o', puzzle.LetterExact},
		// 			{'r', puzzle.LetterExact},
		// 		},
		// 	}},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := evaluateGuessedWord(tt.args.guessedWord, tt.args.solutionWord); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("evaluateGuessedWord() = %v, want %v", got, tt.want)
			}
		})
	}
}

// todo: test for ???:
//   files, err := getAllFilenames(staticFS)
//   log.Printf("  debug fsys:\n    %v\n    %s\n", files, err)

func TestTemplateDataSuggest_validate(t *testing.T) {
	type fields struct {
		Word     string
		Action   string
		Language string
		Message  string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{name: "Suggested word match", fields: fields{Word: "gamer", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "GAMER", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "preuß", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "höste", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "hÖste", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "HÖSTE", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "fülle", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "FÜLLE", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "größe", Action: "add", Language: "english", Message: "test"}, wantErr: nil},
		{name: "Suggested word match", fields: fields{Word: "GRÖßE", Action: "add", Language: "english", Message: "test"}, wantErr: nil},

		{name: "Suggested word invalid (special chars: ?)", fields: fields{Word: "?????"}, wantErr: ErrFailedWordValidation},
		{name: "Suggested word invalid (special chars: ô)", fields: fields{Word: "grôss"}, wantErr: ErrFailedWordValidation},
		{name: "Suggested word invalid (special chars: emoji's (😁))", fields: fields{Word: "😁,😁,😁"}, wantErr: ErrFailedWordValidation},
		{name: "Suggested word invalid (word to short en)", fields: fields{Word: "tiny"}, wantErr: ErrFailedWordValidation},
		{name: "Suggested word invalid (word to short de)", fields: fields{Word: "kurz"}, wantErr: ErrFailedWordValidation},
		{name: "Suggested word invalid (word to long en)", fields: fields{Word: "toolong"}, wantErr: ErrFailedWordValidation},
		{name: "Suggested word invalid (word to long de)", fields: fields{Word: "zulang"}, wantErr: ErrFailedWordValidation},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tds := TemplateDataSuggest{
				Word:     tt.fields.Word,
				Action:   tt.fields.Action,
				Language: tt.fields.Language,
				Message:  tt.fields.Message,
			}
			if err := tds.validate(); err != tt.wantErr {
				t.Errorf("TemplateDataSuggest.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_ExpectedEmbededFiles(t *testing.T) {
	expectedFiles := []string{
		"web/static/generated/main.js",
		"web/static/generated/output.css",
	}

	embededFiles, err := getAllFilenames(fs)
	if err != nil {
		t.Errorf("getAllFilenames() error = %v", err)
	}

	for _, expectedFile := range expectedFiles {
		if !slices.Contains(embededFiles, expectedFile) {
			t.Errorf("expected embeded files, got '%v', want '%v' but was not found", embededFiles, expectedFile)
		}
	}

}

func getAllFilenames(efs iofs.FS) (files []string, err error) {
	if err := iofs.WalkDir(efs, ".", func(path string, d iofs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		files = append(files, path)

		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}
