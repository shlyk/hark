package cmd

import (
	"strings"
	"testing"
)

func TestAskPrintsAnswer(t *testing.T) {
	f := withFakes(t)
	f.output = []byte("B\n")
	out, err := runCapture(t, f, "ask", "Which one?", "--options", "A,B,C")
	if err != nil {
		t.Fatalf("ask error = %v", err)
	}
	if out != "B\n" {
		t.Errorf("ask output = %q, want B\\n", out)
	}
	last := f.runs[len(f.runs)-1]
	if last[0] != "osascript" {
		t.Fatalf("expected osascript run, got %v", f.runs)
	}
	joined := strings.Join(last, "\x00")
	for _, want := range []string{"Which one?", "A", "B", "C", "choose from list"} {
		if !strings.Contains(joined, want) {
			t.Errorf("osascript args missing %q: %v", want, last)
		}
	}
}

func TestAskRecordsHistory(t *testing.T) {
	f := withFakes(t)
	f.output = []byte("Yes\n")
	if _, err := runCapture(t, f, "ask", "Deploy now?"); err != nil {
		t.Fatalf("ask error = %v", err)
	}
	out, err := runCapture(t, f, "history")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Deploy now?") || !strings.Contains(out, "[ask]") {
		t.Errorf("history missing ask entry:\n%s", out)
	}
}

func TestAskFailsWhenCancelled(t *testing.T) {
	f := withFakes(t)
	f.failName = "osascript"
	if _, err := runCapture(t, f, "ask", "Q?"); err == nil {
		t.Error("ask should fail when dialog is cancelled")
	}
}
