package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runWithStdin(t *testing.T, f *fakeExecer, stdin string, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := newRootCmd(f)
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestRelayNotificationPayload(t *testing.T) {
	f := withFakes(t)
	payload := `{"hook_event_name":"Notification","message":"Claude needs permission to run rm"}`
	if _, err := runWithStdin(t, f, payload, "relay", "claude"); err != nil {
		t.Fatalf("relay error = %v", err)
	}
	if len(f.runs) == 0 || f.runs[0][0] != "osascript" {
		t.Fatalf("runs = %v, want osascript banner", f.runs)
	}
	joined := strings.Join(f.runs[0], "\x00")
	if !strings.Contains(joined, "Claude needs permission to run rm") || !strings.Contains(joined, "claude") {
		t.Errorf("banner args missing message/title: %v", f.runs[0])
	}
}

func TestRelayStopPayloadHasDefaultMessage(t *testing.T) {
	f := withFakes(t)
	if _, err := runWithStdin(t, f, `{"hook_event_name":"Stop"}`, "relay", "claude"); err != nil {
		t.Fatalf("relay error = %v", err)
	}
	joined := strings.Join(f.runs[0], "\x00")
	if !strings.Contains(joined, "finished") {
		t.Errorf("Stop relay should use a default 'finished' message: %v", f.runs[0])
	}
}

func TestRelayGarbageStdinFails(t *testing.T) {
	f := withFakes(t)
	if _, err := runWithStdin(t, f, "not json", "relay", "claude"); err == nil {
		t.Error("relay with garbage stdin should fail")
	}
}

func TestRelayEmptyStdinUsesDefaultMessage(t *testing.T) {
	f := withFakes(t)
	if _, err := runWithStdin(t, f, "", "relay", "claude"); err != nil {
		t.Fatalf("relay with empty stdin should still notify, got %v", err)
	}
	joined := strings.Join(f.runs[0], "\x00")
	if !strings.Contains(joined, "needs your attention") {
		t.Errorf("expected default message, args = %v", f.runs[0])
	}
}

func TestHookClaudeSurvivesNullSettings(t *testing.T) {
	f := withFakes(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	os.MkdirAll(filepath.Join(home, ".claude"), 0o755)
	os.WriteFile(settingsPath(home), []byte("null"), 0o644)
	if err := run(t, f, "hook", "claude"); err != nil {
		t.Fatalf("hook claude on null settings error = %v", err)
	}
	data, _ := os.ReadFile(settingsPath(home))
	if !strings.Contains(string(data), "hark relay claude") {
		t.Errorf("hooks not installed over null settings: %s", data)
	}
}

func TestHookRemoveKeepsOtherHooksInSharedGroup(t *testing.T) {
	f := withFakes(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	os.MkdirAll(filepath.Join(home, ".claude"), 0o755)
	// A user manually merged the relay into a group that also runs another hook.
	shared := `{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"hark relay claude"},{"type":"command","command":"my-other-hook"}]}]}}`
	os.WriteFile(settingsPath(home), []byte(shared), 0o644)
	if err := run(t, f, "hook", "claude", "--remove"); err != nil {
		t.Fatalf("hook --remove error = %v", err)
	}
	data, _ := os.ReadFile(settingsPath(home))
	if strings.Contains(string(data), "hark relay claude") {
		t.Errorf("relay hook not removed: %s", data)
	}
	if !strings.Contains(string(data), "my-other-hook") {
		t.Errorf("unrelated hook in the same group was deleted: %s", data)
	}
}

func settingsPath(home string) string {
	return filepath.Join(home, ".claude", "settings.json")
}

func readSettings(t *testing.T, home string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(settingsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestHookClaudeInstallsAndIsIdempotent(t *testing.T) {
	f := withFakes(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	os.MkdirAll(filepath.Join(home, ".claude"), 0o755)
	os.WriteFile(settingsPath(home), []byte(`{"model":"opus"}`), 0o644)

	for i := 0; i < 2; i++ { // run twice: must not duplicate
		if err := run(t, f, "hook", "claude"); err != nil {
			t.Fatalf("hook claude error = %v", err)
		}
	}
	m := readSettings(t, home)
	if m["model"] != "opus" {
		t.Error("existing settings keys must be preserved")
	}
	data, _ := json.Marshal(m)
	if got := strings.Count(string(data), "hark relay claude"); got != 2 { // Notification + Stop
		t.Errorf("expected exactly 2 hook entries, found %d in %s", got, data)
	}
	if _, err := os.Stat(settingsPath(home) + ".bak"); err != nil {
		t.Errorf("backup not created: %v", err)
	}
}

func TestHookClaudeRemove(t *testing.T) {
	f := withFakes(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := run(t, f, "hook", "claude"); err != nil {
		t.Fatal(err)
	}
	if err := run(t, f, "hook", "claude", "--remove"); err != nil {
		t.Fatalf("hook claude --remove error = %v", err)
	}
	data, _ := os.ReadFile(settingsPath(home))
	if strings.Contains(string(data), "hark relay claude") {
		t.Errorf("hooks not removed: %s", data)
	}
}
