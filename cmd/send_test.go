package cmd

import (
	"strings"
	"testing"

	"github.com/shlyk/hark/internal/notify"
)

type fakeExecer struct {
	runs [][]string
}

func (f *fakeExecer) Run(name string, args ...string) error {
	f.runs = append(f.runs, append([]string{name}, args...))
	return nil
}

func (f *fakeExecer) LookPath(name string) (string, error) { return "/usr/bin/" + name, nil }

// withFakes swaps in a fake execer and a temp history path for one test.
func withFakes(t *testing.T) *fakeExecer {
	t.Helper()
	f := &fakeExecer{}
	orig := execer
	execer = f
	t.Cleanup(func() { execer = orig })
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	return f
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	cmd := newRootCmd()
	cmd.SetArgs(args)
	return cmd.Execute()
}

func TestSendJoinsPositionalArgs(t *testing.T) {
	f := withFakes(t)
	if err := run(t, "send", "build", "finished"); err != nil {
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

func TestSendWithSayAlsoSpeaks(t *testing.T) {
	f := withFakes(t)
	if err := run(t, "send", "hello", "--say"); err != nil {
		t.Fatalf("send --say error = %v", err)
	}
	if len(f.runs) != 2 || f.runs[0][0] != "osascript" || f.runs[1][0] != "say" {
		t.Fatalf("runs = %v, want osascript then say", f.runs)
	}
}

func TestSayCommandRunsSay(t *testing.T) {
	f := withFakes(t)
	if err := run(t, "say", "hello", "world", "--voice", "Samantha"); err != nil {
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
