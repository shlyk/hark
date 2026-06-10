package history

import (
	"os"
	"path/filepath"
	"strings"
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

func TestTailHandlesLongLines(t *testing.T) {
	s := tempStore(t)
	long := strings.Repeat("x", 200*1024)
	if err := s.Append(Entry{Time: time.Now(), Kind: "send", Message: long}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Tail(10)
	if err != nil {
		t.Fatalf("Tail() with a 200KB line error = %v", err)
	}
	if len(got) != 1 || got[0].Message != long {
		t.Errorf("Tail() did not return the long entry intact (got %d entries)", len(got))
	}
}

func TestAppendCreatesPrivateFiles(t *testing.T) {
	s := tempStore(t)
	if err := s.Append(Entry{Time: time.Now(), Kind: "send", Message: "secret"}); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(s.Path)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("history file mode = %o, want 600", fi.Mode().Perm())
	}
	di, err := os.Stat(filepath.Dir(s.Path))
	if err != nil {
		t.Fatal(err)
	}
	if di.Mode().Perm() != 0o700 {
		t.Errorf("history dir mode = %o, want 700", di.Mode().Perm())
	}
}

func TestHasRecent(t *testing.T) {
	s := tempStore(t)
	old := Entry{Time: time.Now().Add(-time.Hour), Kind: "send", Message: "old", Key: "k1"}
	fresh := Entry{Time: time.Now().Add(-time.Minute), Kind: "send", Message: "new", Key: "k2"}
	for _, e := range []Entry{old, fresh} {
		if err := s.Append(e); err != nil {
			t.Fatal(err)
		}
	}
	since := time.Now().Add(-10 * time.Minute)
	if got, _ := s.HasRecent("k1", since); got {
		t.Error("HasRecent(k1) = true for an hour-old entry")
	}
	if got, _ := s.HasRecent("k2", since); !got {
		t.Error("HasRecent(k2) = false for a minute-old entry")
	}
	if got, _ := s.HasRecent("missing", since); got {
		t.Error("HasRecent(missing) = true")
	}
	if got, _ := s.HasRecent("", since); got {
		t.Error("HasRecent with empty key must be false")
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
