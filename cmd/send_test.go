package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shlyk/hark/internal/notify"
)

type fakeExecer struct {
	runs     [][]string
	failName string // commands with this name fail
	output   []byte // returned by Output
}

func (f *fakeExecer) Run(name string, args ...string) error {
	f.runs = append(f.runs, append([]string{name}, args...))
	if name == f.failName {
		return errors.New(name + " failed")
	}
	return nil
}

func (f *fakeExecer) Output(name string, args ...string) ([]byte, error) {
	f.runs = append(f.runs, append([]string{name}, args...))
	if name == f.failName {
		return nil, errors.New(name + " failed")
	}
	return f.output, nil
}

func (f *fakeExecer) LookPath(name string) (string, error) { return "/usr/bin/" + name, nil }

// withFakes returns a fake execer and points history and config at temp dirs.
func withFakes(t *testing.T) *fakeExecer {
	t.Helper()
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	return &fakeExecer{}
}

// writeConfig points XDG_CONFIG_HOME at a fresh dir containing the given config.
func writeConfig(t *testing.T, data string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	path := filepath.Join(dir, "hark", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
}

func run(t *testing.T, f *fakeExecer, args ...string) error {
	t.Helper()
	cmd := newRootCmd(f)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func TestSendJoinsPositionalArgs(t *testing.T) {
	f := withFakes(t)
	if err := run(t, f, "send", "build", "finished"); err != nil {
		t.Fatalf("send error = %v", err)
	}
	if len(f.runs) != 1 || f.runs[0][0] != "osascript" {
		t.Fatalf("runs = %v, want one osascript call", f.runs)
	}
	joined := strings.Join(f.runs[0], "\x00")
	if !strings.Contains(joined, "build finished") {
		t.Errorf("expected joined message %q in args %v", "build finished", f.runs[0])
	}
}

func TestSendFlagsReachNotification(t *testing.T) {
	f := withFakes(t)
	if err := run(t, f, "send", "msg", "-t", "myTitle", "-s", "mySubtitle", "--sound", "Glass"); err != nil {
		t.Fatalf("send error = %v", err)
	}
	args := f.runs[0]
	for _, want := range []string{"myTitle", "mySubtitle", "Glass"} {
		found := false
		for _, a := range args {
			if a == want {
				found = true
			}
		}
		if !found {
			t.Errorf("flag value %q not passed to osascript, args = %v", want, args)
		}
	}
}

func TestSendWithSayAlsoSpeaks(t *testing.T) {
	f := withFakes(t)
	if err := run(t, f, "send", "hello", "--say"); err != nil {
		t.Fatalf("send --say error = %v", err)
	}
	if len(f.runs) != 2 || f.runs[0][0] != "osascript" || f.runs[1][0] != "say" {
		t.Fatalf("runs = %v, want osascript then say", f.runs)
	}
}

func TestSendRecordsHistoryWhenSayFails(t *testing.T) {
	f := withFakes(t)
	f.failName = "say"
	if err := run(t, f, "send", "delivered anyway", "--say"); err == nil {
		t.Fatal("send --say with failing say should return an error")
	}
	out, err := runCapture(t, f, "history")
	if err != nil {
		t.Fatalf("history error = %v", err)
	}
	if !strings.Contains(out, "delivered anyway") {
		t.Errorf("delivered banner missing from history after say failure:\n%s", out)
	}
}

func smartProfile(name, transport string) []byte {
	return []byte(`{"SPAudioDataType":[{"_items":[{"_name":"` + name + `","coreaudio_default_audio_output_device":"spaudio_yes","coreaudio_device_transport":"` + transport + `"}]}]}`)
}

func TestSendSmartSpeaksWithHeadphones(t *testing.T) {
	f := withFakes(t)
	f.output = smartProfile("AirPods Max", "coreaudio_device_type_bluetooth")
	if err := run(t, f, "send", "hello", "--smart"); err != nil {
		t.Fatalf("send --smart error = %v", err)
	}
	last := f.runs[len(f.runs)-1]
	if last[0] != "say" {
		t.Errorf("with headphones, expected a say call, runs = %v", f.runs)
	}
}

func TestSendSmartStaysSilentWithoutHeadphones(t *testing.T) {
	f := withFakes(t)
	f.output = smartProfile("MacBook Pro Speakers", "coreaudio_device_type_builtin")
	if err := run(t, f, "send", "hello", "--smart"); err != nil {
		t.Fatalf("send --smart error = %v", err)
	}
	for _, r := range f.runs {
		if r[0] == "say" {
			t.Errorf("without headphones, say must not run, runs = %v", f.runs)
		}
	}
}

func TestConfigDefaultsApplyToSend(t *testing.T) {
	f := withFakes(t)
	writeConfig(t, `{"title":"cfg-title","smart":true}`)
	f.output = smartProfile("AirPods Max", "coreaudio_device_type_bluetooth")
	if err := run(t, f, "send", "hello"); err != nil {
		t.Fatalf("send error = %v", err)
	}
	joined := strings.Join(f.runs[0], "\x00")
	if !strings.Contains(joined, "cfg-title") {
		t.Errorf("config title not applied, args = %v", f.runs[0])
	}
	last := f.runs[len(f.runs)-1]
	if last[0] != "say" {
		t.Errorf("config smart=true with headphones should speak, runs = %v", f.runs)
	}
}

func TestFlagOverridesConfig(t *testing.T) {
	f := withFakes(t)
	writeConfig(t, `{"title":"cfg-title"}`)
	if err := run(t, f, "send", "hello", "-t", "flag-title"); err != nil {
		t.Fatalf("send error = %v", err)
	}
	joined := strings.Join(f.runs[0], "\x00")
	if strings.Contains(joined, "cfg-title") || !strings.Contains(joined, "flag-title") {
		t.Errorf("flag should override config title, args = %v", f.runs[0])
	}
}

func TestSendFailsOnMalformedConfig(t *testing.T) {
	f := withFakes(t)
	writeConfig(t, `{broken`)
	if err := run(t, f, "send", "hello"); err == nil {
		t.Error("send with malformed config should fail loudly")
	}
}

func TestSmartFalseFlagOverridesConfig(t *testing.T) {
	f := withFakes(t)
	writeConfig(t, `{"smart":true}`)
	f.output = smartProfile("AirPods Max", "coreaudio_device_type_bluetooth")
	if err := run(t, f, "send", "hello", "--smart=false"); err != nil {
		t.Fatalf("send error = %v", err)
	}
	for _, r := range f.runs {
		if r[0] == "say" {
			t.Errorf("--smart=false must override config smart=true, runs = %v", f.runs)
		}
	}
}

func TestSendOnceSkipsDuplicate(t *testing.T) {
	f := withFakes(t)
	if err := run(t, f, "send", "build done", "--once", "build-42"); err != nil {
		t.Fatalf("first send error = %v", err)
	}
	first := len(f.runs)
	out, err := runCapture(t, f, "send", "build done", "--once", "build-42")
	if err != nil {
		t.Fatalf("duplicate send error = %v", err)
	}
	if len(f.runs) != first {
		t.Errorf("duplicate send ran commands: %v", f.runs[first:])
	}
	if !strings.Contains(out, "skipped") {
		t.Errorf("duplicate send should say skipped, got %q", out)
	}
	if err := run(t, f, "send", "other", "--once", "other-key"); err != nil {
		t.Fatalf("different key send error = %v", err)
	}
	if len(f.runs) == first {
		t.Error("different key should send")
	}
}

func TestSayCommandRunsSay(t *testing.T) {
	f := withFakes(t)
	if err := run(t, f, "say", "hello", "world", "--voice", "Samantha"); err != nil {
		t.Fatalf("say error = %v", err)
	}
	if len(f.runs) != 1 || f.runs[0][0] != "say" {
		t.Fatalf("runs = %v, want one say call", f.runs)
	}
	want := notify.SayArgs(notify.Speech{Text: "hello world", Voice: "Samantha"})
	got := f.runs[0][1:]
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Errorf("say args = %v, want %v", got, want)
	}
}
