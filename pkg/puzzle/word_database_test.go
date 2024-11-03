package puzzle

import (
	iofs "io/fs"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/pandorasNox/lettr/pkg/language"
)

func TestWordDatabase_Init(t *testing.T) {
	all := `# metadata
gamer
games
`
	common := `# metadata
cried
`
	mockFs := fstest.MapFS{
		"all.txt": {
			// Data: []byte("hello, world"),
			Data: []byte(all),
		},
		"common.txt": {
			Data: []byte(common),
		},
	}

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
			name: "all collection contains common word",
			args: args{
				fs: mockFs,
				filePathsByLanguage: map[language.Language]map[WordCollection][]string{
					language.LANG_EN: map[WordCollection][]string{
						WC_ALL: []string{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wdb := WordDatabase{}
			if err := wdb.Init(tt.args.fs, tt.args.filePathsByLanguage); (err != nil) != tt.wantErr {
				t.Errorf("WordDatabase.Init() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(wdb, tt.wantWdb) {
				t.Errorf("WordDatabase.Init() databases not equal, got %v, want %v", wdb, tt.wantWdb)
			}
		})
	}
}
