package session

import (
	"fmt"
	"log"
	"math/rand"
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

type session struct {
	// todo: think about using mutex or channel for rw session
	id                               string
	expiresAt                        time.Time
	maxAgeSeconds                    int
	language                         language.Language
	gameState                        puzzle.GameState
	pastWords                        []puzzle.Word
	securityHoneypotMessageInputName string
}

func (s *session) AddPastWord(w puzzle.Word) {
	s.pastWords = append(s.pastWords, w)
}

func (s *session) PastWords() []puzzle.Word {
	return slices.Clone(s.pastWords)
}

func (s *session) Language() language.Language {
	return s.language
}

func (s *session) SetLanguage(l language.Language) {
	s.language = l
}

func (s *session) NewGame(l language.Language, wdb puzzle.WordDatabase) {
	s.gameState = puzzle.NewGame(l, wdb, s.PastWords())
}

func (s *session) GameState() *puzzle.GameState {
	return &s.gameState
}

func (s *session) SetGameState(g puzzle.GameState) {
	s.gameState = g
}

func (s *session) SecurityHoneypotMessageInputName() string {
	return s.securityHoneypotMessageInputName
}

func (s *session) NewSecurityHoneypotMessageInputName() {
	name, err := randomString(42)
	if err != nil {
		log.Printf("NewSecurityHoneypotMessageInputName: failed generating random string: %s", err)
		name = "123456789"
	}

	s.securityHoneypotMessageInputName = name
}

func randomString(length int) (string, error) {
	b := make([]byte, length+2)
	randsource := rand.NewSource(time.Now().UnixNano())
	randgenerator := rand.New(randsource)
	_, err := randgenerator.Read(b)
	if err != nil {
		return "", fmt.Errorf("randomString: read from rand failed: %s", err)
	}

	return fmt.Sprintf("%x", b)[2 : length+2], nil
}

func HandleSession(w http.ResponseWriter, req *http.Request, sessions *Sessions, wdb puzzle.WordDatabase) session {
	var err error
	var sess session

	cookie, err := req.Cookie(SESSION_COOKIE_NAME)
	if err != nil {
		return newSession(w, sessions, wdb)
	}

	if cookie == nil {
		return newSession(w, sessions, wdb)
	}

	sid := cookie.Value
	i := slices.IndexFunc(sessions.sessions, func(s session) bool {
		return s.id == sid
	})
	if i == -1 {
		return newSession(w, sessions, wdb)
	}

	sess = sessions.sessions[i]

	c := ConstructCookie(sess)
	http.SetCookie(w, &c)

	sess.expiresAt = generateSessionLifetime()
	sessions.sessions[i] = sess

	return sess
}

func newSession(w http.ResponseWriter, sessions *Sessions, wdb puzzle.WordDatabase) session {
	sess := generateSession(language.LANG_EN, wdb)
	sessions.sessions = append(sessions.sessions, sess)
	c := ConstructCookie(sess)
	http.SetCookie(w, &c)

	return sess
}

func ConstructCookie(s session) http.Cookie {
	return http.Cookie{
		Name:     SESSION_COOKIE_NAME,
		Value:    s.id,
		Path:     "/",
		MaxAge:   s.maxAgeSeconds,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
}

func generateSession(lang language.Language, wdb puzzle.WordDatabase) session { //todo: pass it by ref not by copy?
	id := uuid.NewString()
	expiresAt := generateSessionLifetime()

	return session{id, expiresAt, SESSION_MAX_AGE_IN_SECONDS, lang, puzzle.NewGame(lang, wdb, []puzzle.Word{}), []puzzle.Word{}, ""}
}

func generateSessionLifetime() time.Time {
	return time.Now().Add(SESSION_MAX_AGE_IN_SECONDS * time.Second) // todo: 24 hour, sec now only for testing
}
