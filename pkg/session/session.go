package session

import (
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/pandorasNox/lettr/pkg/language"
	"github.com/pandorasNox/lettr/pkg/puzzle"
)

const SESSION_COOKIE_NAME = "session"
const SESSION_MAX_AGE_IN_SECONDS = 24 * 60 * 60

// type handleSess func(w http.ResponseWriter, req *http.Request, sessions *sessions, wdb puzzle.WordDatabase) Session

type Session struct {
	// todo: think about using mutex or channel for rw session
	id                   string
	expiresAt            time.Time
	maxAgeSeconds        int
	language             language.Language
	activeSolutionWord   puzzle.Word
	lastEvaluatedAttempt puzzle.Puzzle
	pastWords            []puzzle.Word
}

func (s *Session) AddPastWord(w puzzle.Word) {
	s.pastWords = append(s.pastWords, w)
}

func (s *Session) PastWords() []puzzle.Word {
	return slices.Clone(s.pastWords)
}

func (s *Session) Language() language.Language {
	return s.language
}

func (s *Session) SetLanguage(l language.Language) {
	s.language = l
}

func (s *Session) ActiveSolutionWord() puzzle.Word {
	return s.activeSolutionWord
}

func (s *Session) SetActiveSolutionWord(w puzzle.Word) {
	s.activeSolutionWord = w
}

func (s *Session) LastEvaluatedAttempt() puzzle.Puzzle {
	return s.lastEvaluatedAttempt
}

func (s *Session) SetLastEvaluatedAttempt(p puzzle.Puzzle) {
	s.lastEvaluatedAttempt = p
}

type sessions []Session

func NewSessions() sessions {
	return sessions{}
}

func (ss sessions) String() string {
	out := ""
	for _, s := range ss {
		out = out + s.id + " " + s.expiresAt.String() + "\n"
	}

	return out
}

func (ss *sessions) UpdateOrSet(sess Session) {
	index := slices.IndexFunc((*ss), func(s Session) bool {
		return s.id == sess.id
	})
	if index == -1 {
		*ss = append(*ss, sess)
		return
	}

	(*ss)[index] = sess
}

func HandleSession(w http.ResponseWriter, req *http.Request, sessions *sessions, wdb puzzle.WordDatabase) Session {
	var err error
	var sess Session

	cookie, err := req.Cookie(SESSION_COOKIE_NAME)
	if err != nil {
		return newSession(w, sessions, wdb)
	}

	if cookie == nil {
		return newSession(w, sessions, wdb)
	}

	sid := cookie.Value
	i := slices.IndexFunc(*sessions, func(s Session) bool {
		return s.id == sid
	})
	if i == -1 {
		return newSession(w, sessions, wdb)
	}

	sess = (*sessions)[i]

	c := constructCookie(sess)
	http.SetCookie(w, &c)

	sess.expiresAt = generateSessionLifetime()
	(*sessions)[i] = sess

	return sess
}

func newSession(w http.ResponseWriter, sessions *sessions, wdb puzzle.WordDatabase) Session {
	sess := generateSession(language.LANG_EN, wdb)
	*sessions = append(*sessions, sess)
	c := constructCookie(sess)
	http.SetCookie(w, &c)

	return sess
}

func constructCookie(s Session) http.Cookie {
	return http.Cookie{
		Name:     SESSION_COOKIE_NAME,
		Value:    s.id,
		Path:     "/",
		MaxAge:   s.maxAgeSeconds,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func generateSession(lang language.Language, wdb puzzle.WordDatabase) Session { //todo: pass it by ref not by copy?
	id := uuid.NewString()
	expiresAt := generateSessionLifetime()
	activeWord, err := wdb.RandomPick(lang, []puzzle.Word{}, 0)
	if err != nil {
		log.Printf("pick random word failed: %s", err)

		activeWord = puzzle.Word{'R', 'O', 'A', 'T', 'E'}.ToLower()
	}

	return Session{id, expiresAt, SESSION_MAX_AGE_IN_SECONDS, lang, activeWord, puzzle.Puzzle{}, []puzzle.Word{}}
}

func generateSessionLifetime() time.Time {
	return time.Now().Add(SESSION_MAX_AGE_IN_SECONDS * time.Second) // todo: 24 hour, sec now only for testing
}
