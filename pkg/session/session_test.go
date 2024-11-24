package session

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/google/uuid"
	"github.com/pandorasNox/lettr/pkg/language"
	"github.com/pandorasNox/lettr/pkg/puzzle"
)

func Test_ConstructCookie(t *testing.T) {
	fixedUuid := "9566c74d-1003-4c4d-bbbb-0407d1e2c649"
	expireDate := time.Date(2024, 02, 27, 0, 0, 0, 0, time.Now().Location())

	type args struct {
		s session
	}
	tests := []struct {
		name string
		args args
		want http.Cookie
	}{
		// add test cases here
		{
			"test_name",
			args{session{fixedUuid, expireDate, SESSION_MAX_AGE_IN_SECONDS, language.LANG_EN, puzzle.NewGame(language.LANG_EN, puzzle.WordDatabase{}, []puzzle.Word{}), []puzzle.Word{}}},
			http.Cookie{
				Name:     SESSION_COOKIE_NAME,
				Value:    fixedUuid,
				Path:     "/",
				MaxAge:   SESSION_MAX_AGE_IN_SECONDS,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConstructCookie(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("constructCookie() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_HandleSession(t *testing.T) {
	type args struct {
		w        http.ResponseWriter
		req      *http.Request
		sessions *Sessions
		wdb      puzzle.WordDatabase
	}

	// monkey patch time.Now
	patchFnTime := func() time.Time {
		return time.Unix(1615256178, 0)
	}
	patchesTime := gomonkey.ApplyFunc(time.Now, patchFnTime)
	defer patchesTime.Reset()
	// monkey patch uuid.NewString
	patches := gomonkey.ApplyFuncReturn(uuid.NewString, "12345678-abcd-1234-abcd-ab1234567890")
	defer patches.Reset()

	mockWordDatabase := puzzle.WordDatabase{Db: map[language.Language]map[puzzle.WordCollection]map[puzzle.Word]bool{
		language.LANG_EN: {
			puzzle.WC_COMMON: {
				puzzle.Word{'R', 'O', 'A', 'T', 'E'}: true,
			},
		},
	}}

	tests := []struct {
		name string
		args args
		want session
	}{
		// add test cases here
		{
			"test handleSession is generating new session if no cookie is set",
			args{
				httptest.NewRecorder(),
				httptest.NewRequest("get", "/", strings.NewReader("Hello, Reader!")),
				&Sessions{},
				mockWordDatabase,
			},
			session{
				id:            "12345678-abcd-1234-abcd-ab1234567890",
				expiresAt:     time.Unix(1615256178, 0).Add(SESSION_MAX_AGE_IN_SECONDS * time.Second),
				maxAgeSeconds: 86400,
				language:      language.LANG_EN,
				gameState: puzzle.NewGame(
					language.LANG_EN,
					mockWordDatabase,
					[]puzzle.Word{},
				),
				pastWords: []puzzle.Word{},
			},
		},
		// {
		// 	// todo // check out https://gist.github.com/jonnyreeves/17f91155a0d4a5d296d6 for inspiration
		// 	"test got cookie but no session corresponding session on server",
		// 	args{},
		// 	session{
		// 		id:            "12345678-abcd-1234-abcd-ab1234567890",
		// 		expiresAt:     time.Unix(1615256178, 0).Add(SESSION_MAX_AGE_IN_SECONDS * time.Second),
		// 		maxAgeSeconds: 120,
		// 		activeWord:    word{'R','O','A','T','E'},
		// 	},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// fmt.Printf("len=%d cap=%d %v\n", len(tt.want.letterHints), cap(tt.want.letterHints), tt.want.letterHints)
			// got := HandleSession(tt.args.w, tt.args.req, tt.args.sessions, tt.args.wdb)
			// fmt.Printf("len=%d cap=%d %v\n", len(got.letterHints), cap(got.letterHints), got.letterHints)
			if got := HandleSession(tt.args.w, tt.args.req, tt.args.sessions, tt.args.wdb); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleSession() = %v, want %v", got, tt.want)
			}
		})
	}

	// fmt.Println("")
	// fmt.Println("foooooooooooooooo")
	// fmt.Println("")

	// t.Run("test", func(t *testing.T) {
	// 	// t.Errorf("fail %v", session{})
	// 	t.Errorf("fail %v", handleSession(httptest.NewRecorder(), httptest.NewRequest("get", "/", strings.NewReader("Hello, Reader!")), &sessions{}))
	// })
}

func TestSessions_UpdateOrSet(t *testing.T) {
	type args struct {
		sess session
	}
	tests := []struct {
		name string
		ss   *Sessions
		args args
		want Sessions
	}{
		{
			"set new session",
			&Sessions{},
			args{session{id: "foo"}},
			Sessions{
				sessions: []session{
					{id: "foo"},
				},
			},
		},
		{
			"update session",
			&Sessions{[]session{{id: "foo", maxAgeSeconds: 1}}},
			args{session{id: "foo", maxAgeSeconds: 2}},
			Sessions{[]session{{id: "foo", maxAgeSeconds: 2}}},
		},
		{
			"update session changes only correct session",
			&Sessions{[]session{{id: "foo"}, {id: "bar"}, {id: "baz", maxAgeSeconds: 1}, {id: "foobar"}}},
			args{session{id: "baz", maxAgeSeconds: 2}},
			Sessions{[]session{{id: "foo"}, {id: "bar"}, {id: "baz", maxAgeSeconds: 2}, {id: "foobar"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ss.UpdateOrSet(tt.args.sess)
			if !reflect.DeepEqual((*tt.ss), tt.want) {
				t.Errorf("evaluateGuessedWord() = %v, want %v", tt.ss, tt.want)
			}
		})
	}
}
