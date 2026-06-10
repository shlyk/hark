package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/shlyk/hark/internal/notify"
)

type fakeExecer struct {
	runs     [][]string
	failName string // commands with this name fail
}

func (f *fakeExecer) Run(name string, args ...string) error {
	f.runs = append(f.runs, append([]string{name}, args...))
	if name == f.failName {
		return errors.New(name + " failed")
	}
	return nil
}

func (f *fakeExecer) LookPath(name string) (string, error) { return "/usr/bin/" + name, nil }

// withFakes returns a fake execer and points history at a temp dir.
func withFakes(t *testing.T) *fakeExecer {
	t.Helper()
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	return &fakeExecer{}
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
