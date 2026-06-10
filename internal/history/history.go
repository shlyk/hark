// Package history records sent notifications as JSON lines.
package history

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Entry is one recorded notification. Key is the optional dedupe key from
// `send --once`.
type Entry struct {
	Time    time.Time `json:"time"`
	Kind    string    `json:"kind"`
	Title   string    `json:"title,omitempty"`
	Message string    `json:"message"`
	Key     string    `json:"key,omitempty"`
}

// Store reads and appends a JSONL history file.
type Store struct{ Path string }

// DefaultStore resolves $XDG_STATE_HOME/hark/history.jsonl, defaulting to
// ~/.local/state/hark/history.jsonl.
func DefaultStore() (*Store, error) {
	root := os.Getenv("XDG_STATE_HOME")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		root = filepath.Join(home, ".local", "state")
	}
	return &Store{Path: filepath.Join(root, "hark", "history.jsonl")}, nil
}

// Append writes one entry. The file and directory are user-only: history
// records everything agents told the user, which may be sensitive.
func (s *Store) Append(e Entry) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(s.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = f.Write(append(line, '\n'))
	return err
}

// HasRecent reports whether an entry with the given dedupe key exists at or
// after since. The empty key never matches.
func (s *Store) HasRecent(key string, since time.Time) (bool, error) {
	if key == "" {
		return false, nil
	}
	entries, err := s.Tail(0)
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if e.Key == key && !e.Time.Before(since) {
			return true, nil
		}
	}
	return false, nil
}

// Tail returns up to n most recent entries (all if n <= 0). A missing
// history file yields an empty result, not an error. Malformed lines are
// skipped.
func (s *Store) Tail(n int) ([]Entry, error) {
	f, err := os.Open(s.Path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var entries []Entry
	sc := bufio.NewScanner(f)
	// Lift the 64KiB default token limit; agents can log very long messages.
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for sc.Scan() {
		var e Entry
		if json.Unmarshal(sc.Bytes(), &e) == nil {
			entries = append(entries, e)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if n > 0 && len(entries) > n {
		entries = entries[len(entries)-n:]
	}
	return entries, nil
}
