package puzzle

import (
	"reflect"
	"testing"
)

func Test_EvaluateGuessedWord(t *testing.T) {
	type args struct {
		guessedWord  Word
		solutionWord Word
	}
	tests := []struct {
		name string
		args args
		want WordGuess
	}{
		// test cases
		{
			name: "no hits, neither same or exact",
			args: args{
				guessedWord:  Word{},
				solutionWord: Word{'M', 'I', 'S', 'S', 'S'},
			},
			want: WordGuess{
				{Match: MatchNone},
				{Match: MatchNone},
				{Match: MatchNone},
				{Match: MatchNone},
				{Match: MatchNone},
			},
		},
		{
			name: "full exact match",
			args: args{
				guessedWord:  Word{'m', 'a', 't', 'c', 'h'},
				solutionWord: Word{'M', 'A', 'T', 'C', 'H'},
			},
			want: WordGuess{
				{Letter: 'm', Match: MatchExact},
				{Letter: 'a', Match: MatchExact},
				{Letter: 't', Match: MatchExact},
				{Letter: 'c', Match: MatchExact},
				{Letter: 'h', Match: MatchExact},
			},
		},
		{
			name: "partial exact and partial some match",
			args: args{
				guessedWord:  Word{'r', 'a', 'u', 'l', 'o'},
				solutionWord: Word{'R', 'O', 'A', 'T', 'E'},
			},
			want: WordGuess{
				{Letter: 'r', Match: MatchExact},
				{Letter: 'a', Match: MatchVague},
				{Letter: 'u', Match: MatchNone},
				{Letter: 'l', Match: MatchNone},
				{Letter: 'o', Match: MatchVague},
			},
		},
		{
			name: "guessed word contains duplicats",
			args: args{
				guessedWord:  Word{'r', 'o', 't', 'o', 'r'},
				solutionWord: Word{'R', 'O', 'A', 'T', 'E'},
			},
			want: WordGuess{
				{Letter: 'r', Match: MatchExact},
				{Letter: 'o', Match: MatchExact},
				{Letter: 't', Match: MatchVague},
				{Letter: 'o', Match: MatchNone}, // both false bec we already found it or even already guesst the exact match
				{Letter: 'r', Match: MatchNone}, // both false bec we already found it or even already guesst the exact match
			},
		},
		{
			name: "guessed word contains duplicats at end",
			args: args{
				guessedWord:  Word{'i', 'x', 'i', 'i', 'i'},
				solutionWord: Word{'L', 'X', 'I', 'I', 'I'},
			},
			want: WordGuess{
				{Letter: 'i', Match: MatchNone},
				{Letter: 'x', Match: MatchExact},
				{Letter: 'i', Match: MatchExact},
				{Letter: 'i', Match: MatchExact},
				{Letter: 'i', Match: MatchExact},
			},
		},
		{
			name: "guessed word contains duplicats at end fpp",
			args: args{
				guessedWord:  Word{'l', 'i', 'i', 'i', 'i'},
				solutionWord: Word{'I', 'L', 'X', 'I', 'I'},
			},
			want: WordGuess{
				{Letter: 'l', Match: MatchVague},
				{Letter: 'i', Match: MatchVague},
				{Letter: 'i', Match: MatchNone},
				{Letter: 'i', Match: MatchExact},
				{Letter: 'i', Match: MatchExact},
			},
		},
		// {
		// 	name: "target word contains duplicats / guessed word contains duplicats",
		// 	args: args{
		// 		Puzzle{},
		// 		url.Values{"r0c0": []string{"M"}, "r0c1": []string{"A"}, "r0c2": []string{"T"}, "r0c3": []string{"C"}, "r0c4": []string{"H"}},
		// 		word{'M', 'A', 'T', 'C', 'H'},
		// 	},
		// 	want: Puzzle{"", wordGuess{
		// 		{
		// 			{'r', LetterExact},
		// 			{'o', LetterExact},
		// 			{'t', LetterExact},
		// 			{'o', LetterExact},
		// 			{'r', LetterExact},
		// 		},
		// 	}},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EvaluateGuessedWord(tt.args.guessedWord, tt.args.solutionWord); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("evaluateGuessedWord() = %v, want %v", got, tt.want)
			}
		})
	}
}
