// Package session tracks portfolio visit statistics: a persisted total-visitor
// count and per-username last-login times, plus a live active-session gauge.
package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type state struct {
	TotalVisitors int64                `json:"totalVisitors"`
	LastLogins    map[string]time.Time `json:"lastLogins"`
}

// Store is a concurrency-safe, best-effort persistent visitor store.
type Store struct {
	path      string
	mu        sync.Mutex
	st        state
	active    int64
	startedAt time.Time
}

// NewStore loads existing state from path (if present) and returns a Store.
// Persistence is best-effort: failures never block a session.
func NewStore(path string) *Store {
	s := &Store{path: path, startedAt: time.Now(), st: state{LastLogins: map[string]time.Time{}}}
	if b, err := os.ReadFile(path); err == nil {
		var loaded state
		if json.Unmarshal(b, &loaded) == nil {
			if loaded.LastLogins == nil {
				loaded.LastLogins = map[string]time.Time{}
			}
			s.st = loaded
		}
	}
	return s
}

// Visit records a login for user, returning their previous last-login time and
// whether this is their first-ever visit.
func (s *Store) Visit(user string) (prev time.Time, first bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prev = s.st.LastLogins[user]
	first = prev.IsZero()
	s.st.TotalVisitors++
	s.st.LastLogins[user] = time.Now()
	s.save()
	return prev, first
}

// save writes state to disk; callers must hold s.mu.
func (s *Store) save() {
	if s.path == "" {
		return
	}
	if dir := filepath.Dir(s.path); dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	if b, err := json.MarshalIndent(s.st, "", "  "); err == nil {
		tmp := s.path + ".tmp"
		if os.WriteFile(tmp, b, 0o644) == nil {
			_ = os.Rename(tmp, s.path)
		}
	}
}

func (s *Store) IncActive() { atomic.AddInt64(&s.active, 1) }
func (s *Store) DecActive() { atomic.AddInt64(&s.active, -1) }
func (s *Store) Active() int64 {
	return atomic.LoadInt64(&s.active)
}

func (s *Store) Total() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.st.TotalVisitors
}

func (s *Store) Uptime() time.Duration { return time.Since(s.startedAt) }
