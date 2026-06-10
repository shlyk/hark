// Package history records sent notifications as JSON lines.
package history

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	Time    time.Time `json:"time"`
	Kind    string    `json:"kind"`
	Title   string    `json:"title,omitempty"`
	Message string    `json:"message"`
}

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

func (s *Store) Append(e Entry) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(s.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
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
