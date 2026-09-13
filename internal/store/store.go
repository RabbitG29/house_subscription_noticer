package store

import (
	"encoding/json"
	"os"
	"sync"
)

// Store tracks which announcement IDs we've already notified about,
// persisted as a plain JSON file so the bot remembers state across
// restarts without needing an external database dependency (DB가 필요해지면
// 이 구현체만 SQLite 등으로 교체하면 됩니다).
type Store struct {
	path string
	mu   sync.Mutex
	seen map[string]bool
}

func Load(path string) (*Store, error) {
	s := &Store{path: path, seen: make(map[string]bool)}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(data, &s.seen); err != nil {
		return nil, err
	}
	return s, nil
}

// Has reports whether id has already been recorded, without mutating state.
// Callers should only record an id (via MarkSeen) once the notification for
// it has actually been delivered — recording it earlier would mean a failed
// send (e.g. SMTP outage) is never retried.
func (s *Store) Has(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.seen[id]
}

// MarkSeen records id as notified (in memory — call Save to persist to disk).
func (s *Store) MarkSeen(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen[id] = true
}

func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(s.seen, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}
