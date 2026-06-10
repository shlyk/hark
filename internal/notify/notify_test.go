package notify

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type fakeExecer struct {
	lookPathErr error
	runErr      error
	ranName     string
	ranArgs     []string
}

func (f *fakeExecer) Run(name string, args ...string) error {
	f.ranName = name
	f.ranArgs = args
	return f.runErr
}

func (f *fakeExecer) LookPath(name string) (string, error) {
	return "/usr/bin/" + name, f.lookPathErr
}

func TestScriptNeverContainsUserText(t *testing.T) {
	n := Notification{Message: `"' & do shell script "rm -rf ~"`, Title: "evil", Subtitle: "x", Sound: "Glass"}
	script := Script(n)
	for _, s := range []string{n.Message, n.Title, n.Subtitle, n.Sound} {
		if strings.Contains(script, s) {
			t.Errorf("script must not interpolate user text, found %q in:\n%s", s, script)
		}
	}
}

func TestScriptIncludesOnlySetClauses(t *testing.T) {
	bare := Script(Notification{Message: "m", Title: "t"})
	if strings.Contains(bare, "subtitle") || strings.Contains(bare, "sound name") {
		t.Errorf("bare script must not contain subtitle/sound clauses:\n%s", bare)
	}
	full := Script(Notification{Message: "m", Title: "t", Subtitle: "s", Sound: "Glass"})
	if !strings.Contains(full, "subtitle (item 3 of argv)") || !strings.Contains(full, "sound name (item 4 of argv)") {
		t.Errorf("full script missing subtitle/sound clauses:\n%s", full)
	}
}

func TestArgsPassValuesAsArgv(t *testing.T) {
	n := Notification{Message: "msg", Title: "ti", Subtitle: "su", Sound: "Glass"}
	args := Args(n)
	want := []string{"-e", Script(n), "--", "msg", "ti", "su", "Glass"}
	if !reflect.DeepEqual(args, want) {
		t.Errorf("Args() = %v, want %v", args, want)
	}
}

func TestSendRunsOsascript(t *testing.T) {
	f := &fakeExecer{}
	err := Send(f, Notification{Message: "hello", Title: "hark"})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if f.ranName != "osascript" {
		t.Errorf("ran %q, want osascript", f.ranName)
	}
}

func TestSendRejectsEmptyMessage(t *testing.T) {
	f := &fakeExecer{}
	if err := Send(f, Notification{Message: "   "}); err == nil {
		t.Error("Send() with blank message should error")
	}
	if f.ranName != "" {
		t.Error("Send() must not run anything for blank message")
	}
}

func TestSendFailsWhenOsascriptMissing(t *testing.T) {
	f := &fakeExecer{lookPathErr: errors.New("not found")}
	if err := Send(f, Notification{Message: "hi"}); err == nil {
		t.Error("Send() should fail when osascript is missing")
	}
}

func TestSendWrapsRunError(t *testing.T) {
	f := &fakeExecer{runErr: errors.New("boom")}
	if err := Send(f, Notification{Message: "hi"}); err == nil {
		t.Error("Send() should fail when osascript exits non-zero")
	}
}

func TestSayArgs(t *testing.T) {
	got := SayArgs(Speech{Text: "hello there", Voice: "Samantha", Rate: 200})
	want := []string{"-v", "Samantha", "-r", "200", "--", "hello there"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SayArgs() = %v, want %v", got, want)
	}
	gotBare := SayArgs(Speech{Text: "hi"})
	wantBare := []string{"--", "hi"}
	if !reflect.DeepEqual(gotBare, wantBare) {
		t.Errorf("SayArgs() bare = %v, want %v", gotBare, wantBare)
	}
}

func TestSayRejectsNegativeRate(t *testing.T) {
	f := &fakeExecer{}
	if err := Say(f, Speech{Text: "hi", Rate: -50}); err == nil {
		t.Error("Say() with negative rate should error")
	}
}

func TestSayRejectsEmptyText(t *testing.T) {
	f := &fakeExecer{}
	if err := Say(f, Speech{Text: ""}); err == nil {
		t.Error("Say() with empty text should error")
	}
}
