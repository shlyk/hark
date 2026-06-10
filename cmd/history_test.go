package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

func runCapture(t *testing.T, f *fakeExecer, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := newRootCmd(f)
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestHistoryShowsSentNotifications(t *testing.T) {
	f := withFakes(t)
	if err := run(t, f, "send", "first"); err != nil {
		t.Fatal(err)
	}
	if err := run(t, f, "send", "second"); err != nil {
		t.Fatal(err)
	}
	out, err := runCapture(t, f, "history")
	if err != nil {
		t.Fatalf("history error = %v", err)
	}
	if !strings.Contains(out, "first") || !strings.Contains(out, "second") {
		t.Errorf("history output missing entries:\n%s", out)
	}
}

func TestHistoryJSON(t *testing.T) {
	f := withFakes(t)
	if err := run(t, f, "send", "hello"); err != nil {
		t.Fatal(err)
	}
	out, err := runCapture(t, f, "history", "--json")
	if err != nil {
		t.Fatalf("history --json error = %v", err)
	}
	var entries []map[string]any
	if err := json.Unmarshal([]byte(out), &entries); err != nil {
		t.Fatalf("output is not a JSON array: %v\n%s", err, out)
	}
	if len(entries) != 1 || entries[0]["message"] != "hello" {
		t.Errorf("entries = %v, want one entry with message hello", entries)
	}
}

func TestHistoryEmptyIsOK(t *testing.T) {
	f := withFakes(t)
	if _, err := runCapture(t, f, "history"); err != nil {
		t.Errorf("history with no file should succeed, got %v", err)
	}
}

type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func TestHistoryFollowPrintsNewEntries(t *testing.T) {
	f := withFakes(t)
	if err := run(t, f, "send", "first"); err != nil {
		t.Fatal(err)
	}
	orig := followInterval
	followInterval = 10 * time.Millisecond
	t.Cleanup(func() { followInterval = orig })

	ctx, cancel := context.WithCancel(context.Background())
	out := &syncBuffer{}
	cmd := newRootCmd(f)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"history", "--follow"})
	done := make(chan error, 1)
	go func() { done <- cmd.ExecuteContext(ctx) }()

	time.Sleep(50 * time.Millisecond)
	if err := run(t, f, "send", "second"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("history --follow error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "first") || !strings.Contains(got, "second") {
		t.Errorf("follow output missing entries:\n%s", got)
	}
}

func TestHistoryFollowRejectsJSON(t *testing.T) {
	f := withFakes(t)
	if _, err := runCapture(t, f, "history", "--follow", "--json"); err == nil {
		t.Error("history --follow --json should be rejected")
	}
}

func TestDoctorReportsChecks(t *testing.T) {
	f := withFakes(t)
	f.output = []byte(`"HIDIdleTime" = 5000000000`)
	t.Setenv("HOME", t.TempDir())
	out, err := runCapture(t, f, "doctor")
	if err != nil {
		t.Fatalf("doctor error = %v\n%s", err, out)
	}
	for _, want := range []string{"osascript", "say", "history", "test notification", "presence", "ntfy"} {
		if !strings.Contains(out, want) {
			t.Errorf("doctor output missing %q:\n%s", want, out)
		}
	}
}

func TestDoctorFailsOnMalformedConfig(t *testing.T) {
	f := withFakes(t)
	f.output = []byte(`"HIDIdleTime" = 5000000000`)
	t.Setenv("HOME", t.TempDir())
	writeConfig(t, `{broken`)
	out, err := runCapture(t, f, "doctor")
	if err == nil {
		t.Errorf("doctor with malformed config should fail:\n%s", out)
	}
	if !strings.Contains(out, "config") {
		t.Errorf("doctor output should mention config:\n%s", out)
	}
}

func TestDoctorReportsNtfyTopic(t *testing.T) {
	f := withFakes(t)
	f.output = []byte(`"HIDIdleTime" = 5000000000`)
	t.Setenv("HOME", t.TempDir())
	writeConfig(t, `{"ntfy":{"topic":"my-topic"}}`)
	out, err := runCapture(t, f, "doctor")
	if err != nil {
		t.Fatalf("doctor error = %v\n%s", err, out)
	}
	if !strings.Contains(out, "ntfy") || !strings.Contains(out, "configured") {
		t.Errorf("doctor should report ntfy as configured:\n%s", out)
	}
}
