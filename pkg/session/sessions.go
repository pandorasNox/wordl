package session

import (
	"fmt"
	"slices"
	"time"
)

type ISessions interface {
	fmt.Stringer
	UpdateOrSet(session)
	RemoveExpiredSessions()
}

type Sessions struct {
	sessions []session
}

// ensure interface implementation
// var _ Sessioner = Sessions{}
var _ ISessions = (*Sessions)(nil)

func NewSessions() Sessions {
	return Sessions{}
}

func (ss *Sessions) String() string {
	out := ""
	for _, s := range ss.sessions {
		out = out + s.id + " " + s.expiresAt.String() + "\n"
	}

	return out
}

func (s *Sessions) GetById(sid string) (session, error) {
	i := slices.IndexFunc(s.sessions, func(s session) bool {
		return s.id == sid
	})
	if i == -1 {
		return session{}, fmt.Errorf("no session found with id='%s'", sid)
	}

	return s.sessions[i], nil
}

func (ss *Sessions) UpdateOrSet(sess session) {
	index := slices.IndexFunc((ss.sessions), func(s session) bool {
		return s.id == sess.id
	})
	if index == -1 {
		ss.sessions = append(ss.sessions, sess)
		return
	}

	(ss.sessions)[index] = sess
}

func (ss *Sessions) RemoveExpiredSessions() {
	now := time.Now()
	ss.sessions = slices.DeleteFunc(ss.sessions, func(s session) bool {
		return now.After(s.expiresAt)
	})
}
