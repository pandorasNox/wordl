package session

import (
	"reflect"
	"testing"
	"time"
)

func TestSessions_RemoveExpiredSessions(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name           string
		sessionsBefore []session
		sessionsAfter  []session
	}{
		{
			name: "no change",
			sessionsBefore: []session{
				{
					expiresAt: now.Add(1 * time.Hour),
				},
			},
			sessionsAfter: []session{
				{
					expiresAt: now.Add(1 * time.Hour),
				},
			},
		},
		{
			name: "remove multiple expired",
			sessionsBefore: []session{
				{
					expiresAt: now.Add(-2 * time.Hour),
				},
				{
					expiresAt: now.Add(1 * time.Hour),
				},
				{
					expiresAt: now.Add(-1 * time.Hour),
				},
				{
					expiresAt: now.Add(-100 * time.Hour),
				},
			},
			sessionsAfter: []session{
				{
					expiresAt: now.Add(1 * time.Hour),
				},
			},
		},
		{
			name: "remove everything",
			sessionsBefore: []session{
				{
					expiresAt: now.Add(-2 * time.Hour),
				},
				{
					expiresAt: now.Add(-1 * time.Hour),
				},
			},
			sessionsAfter: []session{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &Sessions{
				sessions: tt.sessionsBefore,
			}
			ss.RemoveExpiredSessions()

			if !reflect.DeepEqual((ss.sessions), tt.sessionsAfter) {
				t.Errorf("RemoveExpiredSessions() = %v, want %v", ss.sessions, tt.sessionsAfter)
			}
		})
	}
}
