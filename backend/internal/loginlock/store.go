package loginlock

import (
	"sync"
	"time"
)

// entry holds failed count and expiry for a username.
type entry struct {
	count    int
	expireAt time.Time
}

// Store tracks failed login attempts per username with TTL (e.g. 3600s).
// Safe for concurrent use.
type Store struct {
	mu      sync.Mutex
	entries map[string]*entry
	ttlSec  int
}

// NewStore returns a new in-memory store. ttlSec is the expiry in seconds (e.g. 3600).
func NewStore(ttlSec int) *Store {
	return &Store{entries: make(map[string]*entry), ttlSec: ttlSec}
}

// Add increments failed count for username and sets expiry from now.
func (s *Store) Add(username string) {
	if username == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanLocked()
	e, ok := s.entries[username]
	if !ok || e.expireAt.Before(time.Now()) {
		e = &entry{count: 1, expireAt: time.Now().Add(time.Duration(s.ttlSec) * time.Second)}
		s.entries[username] = e
	} else {
		e.count++
	}
}

// Count returns current failed count for username (0 if expired or not found).
func (s *Store) Count(username string) int {
	if username == "" {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanLocked()
	e, ok := s.entries[username]
	if !ok || e.expireAt.Before(time.Now()) {
		return 0
	}
	return e.count
}

// Clear removes the entry for username (call on successful login).
func (s *Store) Clear(username string) {
	if username == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, username)
}

func (s *Store) cleanLocked() {
	now := time.Now()
	for k, e := range s.entries {
		if e.expireAt.Before(now) {
			delete(s.entries, k)
		}
	}
}
