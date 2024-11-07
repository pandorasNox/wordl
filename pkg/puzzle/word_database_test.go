package puzzle

import (
	"errors"
	iofs "io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"time"

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

type mockFileWithStatError struct{}

func (f mockFileWithStatError) Stat() (iofs.FileInfo, error) {
	return nil, errors.New("")
}
func (f mockFileWithStatError) Read([]byte) (int, error) {
	return -1, errors.New("")
}
func (f mockFileWithStatError) Close() error {
	return errors.New("")
}

type mockFsWithStatErrorFile struct{}

func (fs mockFsWithStatErrorFile) Open(_ string) (iofs.File, error) {
	return mockFileWithStatError{}, nil
}

type mockFileInfoWithBigFile struct{}

func (f mockFileInfoWithBigFile) Name() string {
	return "name"
}

func (f mockFileInfoWithBigFile) Size() int64 {
	return 10 * 1024 * 1024 // 10 mega byte
}

func (f mockFileInfoWithBigFile) Mode() iofs.FileMode {
	return iofs.ModeTemporary
}

func (f mockFileInfoWithBigFile) ModTime() time.Time {
	return time.Now()
}

func (f mockFileInfoWithBigFile) IsDir() bool {
	return false
}

func (f mockFileInfoWithBigFile) Sys() any {
	return nil
}

type mockFileWithBigFile struct{}

func (f mockFileWithBigFile) Stat() (iofs.FileInfo, error) {
	return mockFileInfoWithBigFile{}, nil
}

func (f mockFileWithBigFile) Read([]byte) (int, error) {
	return -1, errors.New("")
}
func (f mockFileWithBigFile) Close() error {
	return errors.New("")
}

type mockFsWithBigFile struct{}

func (fs mockFsWithBigFile) Open(_ string) (iofs.File, error) {
	return mockFileWithBigFile{}, nil
}

func TestWordDatabase_Init(t *testing.T) {
	type args struct {
		fs                  iofs.FS
		filePathsByLanguage map[language.Language]map[WordCollection][]string
	}
	tests := []struct {
		name                   string
		args                   args
		wantErr                bool
		wantErrMessageContains string
		wantWdb                WordDatabase
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
				fs: fstest.MapFS{},
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
			wantErr:                true,
			wantErrMessageContains: "wordDatabase init failed when opening file",
			wantWdb: WordDatabase{
				Db: map[language.Language]map[WordCollection]map[Word]bool{
					language.LANG_EN: {
						WC_ALL: {},
					},
				},
			},
		},
		//
		{
			name: "Init can't get file stat",
			args: args{
				fs: mockFsWithStatErrorFile{},
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
			wantErr:                true,
			wantErrMessageContains: "wordDatabase init failed when obtaining stat",
			wantWdb:                WordDatabase{},
		},
		//
		{
			name: "Init fails due to too big file",
			args: args{
				fs: mockFsWithBigFile{},
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
			wantErr:                true,
			wantErrMessageContains: "wordDatabase init failed with forbidden file size",
			wantWdb:                WordDatabase{},
		},
		//
		{
			name: "init fails due invalid word in file",
			args: args{
				fs: fstest.MapFS{
					"common.txt": {
						Data: []byte(`# metadata
gamer
notalettrword
`),
					},
				},
				filePathsByLanguage: map[language.Language]map[WordCollection][]string{
					language.LANG_EN: {
						WC_COMMON: {
							"common.txt",
						},
					},
				},
			},
			wantErr:                true,
			wantErrMessageContains: "wordDatabase init, couldn't parse line to word",
			wantWdb: WordDatabase{
				Db: map[language.Language]map[WordCollection]map[Word]bool{
					language.LANG_EN: {
						WC_ALL: {
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
			err := wdb.Init(tt.args.fs, tt.args.filePathsByLanguage)
			if (err != nil) != tt.wantErr {
				t.Errorf("WordDatabase.Init() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr && tt.wantErrMessageContains != "" {
				if !strings.Contains(err.Error(), tt.wantErrMessageContains) {
					t.Errorf("WordDatabase.Init() error = %v, wantErr %v, wantErrMessageContains %s", err, tt.wantErr, tt.wantErrMessageContains)
				}
			}
			if tt.wantErr == false && !reflect.DeepEqual(wdb, tt.wantWdb) {
				t.Errorf("WordDatabase.Init() databases not equal, got %v, want %v", wdb, tt.wantWdb)
			}
		})
	}
}
