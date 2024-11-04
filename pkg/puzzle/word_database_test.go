package puzzle

import (
	iofs "io/fs"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/pandorasNox/lettr/pkg/language"
)

// func wordTxtByName(name string) []byte {
// 	switch name {
// 	case "case01_all":
// 		return []byte(`# metadata
// gamer
// games
// `)
// 	case "case01_common":
// 		return []byte(`# metadata
// cried
// `)
// 	default:
// 		return []byte{}
// 	}
// }

func TestWordDatabase_Init(t *testing.T) {
	type args struct {
		fs                  iofs.FS
		filePathsByLanguage map[language.Language]map[WordCollection][]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantWdb WordDatabase
	}{
		// Add test cases.
		{
			name: "allCollection contains commonCollection word",
			args: args{
				fs: fstest.MapFS{
					"all.txt": {
						// Data: []byte("hello, world"),
						Data: []byte(`# metadata
gamer
games
`),
					},
					"common.txt": {
						Data: []byte(`# metadata
cried
`),
					},
				},
				filePathsByLanguage: map[language.Language]map[WordCollection][]string{
					language.LANG_EN: {
						WC_ALL: {
							"all.txt",
						},
						WC_COMMON: {
							"common.txt",
						},
					},
				},
			},
			wantErr: false,
			wantWdb: WordDatabase{
				Db: map[language.Language]map[WordCollection]map[Word]bool{
					language.LANG_EN: {
						WC_ALL: {
							{'g', 'a', 'm', 'e', 's'}: true,
							{'g', 'a', 'm', 'e', 'r'}: true,
							{'c', 'r', 'i', 'e', 'd'}: true,
						},
						WC_COMMON: {
							{'c', 'r', 'i', 'e', 'd'}: true,
						},
					},
				},
			},
		},
		//
		{
			name: "allCollection dedublicates word from commonCollection",
			args: args{
				fs: fstest.MapFS{
					"all.txt": {
						// Data: []byte("hello, world"),
						Data: []byte(`# metadata
gamer
games
`),
					},
					"common.txt": {
						Data: []byte(`# metadata
gamer
`),
					},
				},
				filePathsByLanguage: map[language.Language]map[WordCollection][]string{
					language.LANG_EN: {
						WC_ALL: {
							"all.txt",
						},
						WC_COMMON: {
							"common.txt",
						},
					},
				},
			},
			wantErr: false,
			wantWdb: WordDatabase{
				Db: map[language.Language]map[WordCollection]map[Word]bool{
					language.LANG_EN: {
						WC_ALL: {
							{'g', 'a', 'm', 'e', 's'}: true,
							{'g', 'a', 'm', 'e', 'r'}: true,
						},
						WC_COMMON: {
							{'g', 'a', 'm', 'e', 'r'}: true,
						},
					},
				},
			},
		},
		//
		{
			name: "Init can't open files",
			args: args{
				fs: fstest.MapFS{
					"all.txt": {
						// Data: []byte("hello, world"),
						Data: []byte(`# metadata
gamer
games
`),
					},
					"common.txt": {
						Data: []byte(`# metadata
gamer
`),
					},
				},
				filePathsByLanguage: map[language.Language]map[WordCollection][]string{
					language.LANG_EN: {
						WC_ALL: {
							"non-existing-file-path",
						},
						WC_COMMON: {
							"non-existing-file-path",
						},
					},
				},
			},
			wantErr: true,
			wantWdb: WordDatabase{
				Db: map[language.Language]map[WordCollection]map[Word]bool{
					language.LANG_EN: {
						WC_ALL: {
							{'g', 'a', 'm', 'e', 's'}: true,
							{'g', 'a', 'm', 'e', 'r'}: true,
						},
						WC_COMMON: {
							{'g', 'a', 'm', 'e', 'r'}: true,
						},
					},
				},
			},
		},
		//
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wdb := WordDatabase{}
			if err := wdb.Init(tt.args.fs, tt.args.filePathsByLanguage); (err != nil) != tt.wantErr {
				t.Errorf("WordDatabase.Init() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == false && !reflect.DeepEqual(wdb, tt.wantWdb) {
				t.Errorf("WordDatabase.Init() databases not equal, got %v, want %v", wdb, tt.wantWdb)
			}
		})
	}
}
