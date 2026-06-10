package history

import (
	"path/filepath"
	"testing"
	"time"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	return &Store{Path: filepath.Join(t.TempDir(), "hark", "history.jsonl")}
}

func TestAppendAndTail(t *testing.T) {
	s := tempStore(t)
	e := Entry{Time: time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC), Kind: "send", Title: "CI", Message: "done"}
	if err := s.Append(e); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	got, err := s.Tail(10)
	if err != nil {
		t.Fatalf("Tail() error = %v", err)
	}
	if len(got) != 1 || got[0].Message != "done" || got[0].Kind != "send" || !got[0].Time.Equal(e.Time) {
		t.Errorf("Tail() = %+v, want one entry matching %+v", got, e)
	}
}

func TestTailLimitReturnsMostRecent(t *testing.T) {
	s := tempStore(t)
	for i := 0; i < 5; i++ {
		if err := s.Append(Entry{Time: time.Now(), Kind: "send", Message: string(rune('a' + i))}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.Tail(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Message != "d" || got[1].Message != "e" {
		t.Errorf("Tail(2) = %+v, want last two entries d, e", got)
	}
}

func TestTailMissingFile(t *testing.T) {
	s := tempStore(t)
	got, err := s.Tail(10)
	if err != nil {
		t.Fatalf("Tail() on missing file error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Tail() on missing file = %+v, want empty", got)
	}
}

func TestDefaultStoreUsesXDGStateHome(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/tmp/xdg-state")
	s, err := DefaultStore()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/tmp/xdg-state", "hark", "history.jsonl")
	if s.Path != want {
		t.Errorf("DefaultStore().Path = %q, want %q", s.Path, want)
	}
}
