package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func runCapture(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestHistoryShowsSentNotifications(t *testing.T) {
	withFakes(t)
	if err := run(t, "send", "first"); err != nil {
		t.Fatal(err)
	}
	if err := run(t, "send", "second"); err != nil {
		t.Fatal(err)
	}
	out, err := runCapture(t, "history")
	if err != nil {
		t.Fatalf("history error = %v", err)
	}
	if !strings.Contains(out, "first") || !strings.Contains(out, "second") {
		t.Errorf("history output missing entries:\n%s", out)
	}
}

func TestHistoryJSON(t *testing.T) {
	withFakes(t)
	if err := run(t, "send", "hello"); err != nil {
		t.Fatal(err)
	}
	out, err := runCapture(t, "history", "--json")
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
	withFakes(t)
	if _, err := runCapture(t, "history"); err != nil {
		t.Errorf("history with no file should succeed, got %v", err)
	}
}

func TestDoctorReportsChecks(t *testing.T) {
	withFakes(t)
	out, err := runCapture(t, "doctor")
	if err != nil {
		t.Fatalf("doctor error = %v\n%s", err, out)
	}
	for _, want := range []string{"osascript", "say", "history", "test notification"} {
		if !strings.Contains(out, want) {
			t.Errorf("doctor output missing %q:\n%s", want, out)
		}
	}
}
