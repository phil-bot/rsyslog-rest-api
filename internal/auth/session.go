package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const sessionTTL = 8 * time.Hour

type session struct {
	token     string
	expiresAt time.Time
}

// SessionStore is a simple in-memory store for admin session tokens.
// Tokens are invalidated on restart, which is acceptable for an admin UI.
type SessionStore struct {
	mu       sync.Mutex
	sessions map[string]session
}

// NewSessionStore creates an empty SessionStore.
func NewSessionStore() *SessionStore {
	s := &SessionStore{
		sessions: make(map[string]session),
	}
	go s.cleanupLoop()
	return s
}

// Create generates a new session token and stores it.
func (s *SessionStore) Create() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)

	s.mu.Lock()
	s.sessions[token] = session{
		token:     token,
		expiresAt: time.Now().Add(sessionTTL),
	}
	s.mu.Unlock()

	return token, nil
}

// Validate checks whether a token is valid and not expired.
func (s *SessionStore) Validate(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[token]
	if !ok {
		return false
	}
	if time.Now().After(sess.expiresAt) {
		delete(s.sessions, token)
		return false
	}
	return true
}

// Revoke deletes a session token (logout).
func (s *SessionStore) Revoke(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

// cleanupLoop periodically removes expired sessions.
func (s *SessionStore) cleanupLoop() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for token, sess := range s.sessions {
			if now.After(sess.expiresAt) {
				delete(s.sessions, token)
			}
		}
		s.mu.Unlock()
	}
}
