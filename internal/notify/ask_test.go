package notify

import (
	"errors"
	"strings"
	"testing"
)

func TestAskScriptNeverContainsUserText(t *testing.T) {
	q := Question{Prompt: `evil" & quit`, Title: "t\"", Options: []string{`opt"1`, "opt2"}}
	script := AskScript(q)
	for _, s := range []string{q.Prompt, q.Title, q.Options[0]} {
		if strings.Contains(script, s) {
			t.Errorf("ask script must not interpolate user text, found %q in:\n%s", s, script)
		}
	}
}

func TestAskScriptVariants(t *testing.T) {
	yesno := AskScript(Question{Prompt: "q", Title: "t"})
	if !strings.Contains(yesno, `buttons {"No", "Yes"}`) {
		t.Errorf("yes/no script missing buttons:\n%s", yesno)
	}
	list := AskScript(Question{Prompt: "q", Title: "t", Options: []string{"A", "B"}})
	if !strings.Contains(list, "choose from list") {
		t.Errorf("options script must use choose from list:\n%s", list)
	}
	input := AskScript(Question{Prompt: "q", Title: "t", Input: true})
	if !strings.Contains(input, "default answer") || !strings.Contains(input, "text returned") {
		t.Errorf("input script missing text field handling:\n%s", input)
	}
	timed := AskScript(Question{Prompt: "q", Title: "t", Timeout: 30})
	if !strings.Contains(timed, "giving up after 30") {
		t.Errorf("timeout script missing giving up clause:\n%s", timed)
	}
}

func TestAskArgsPassValuesAsArgv(t *testing.T) {
	q := Question{Prompt: "q", Title: "t", Options: []string{"A", "B"}}
	args := AskArgs(q)
	want := []string{"-e", AskScript(q), "--", "q", "t", "A", "B"}
	if strings.Join(args, "\x00") != strings.Join(want, "\x00") {
		t.Errorf("AskArgs() = %v, want %v", args, want)
	}
}

func TestAskReturnsTrimmedAnswer(t *testing.T) {
	f := &fakeExecer{output: []byte("Yes\n")}
	got, err := Ask(f, Question{Prompt: "q", Title: "t"})
	if err != nil {
		t.Fatalf("Ask() error = %v", err)
	}
	if got != "Yes" {
		t.Errorf("Ask() = %q, want Yes", got)
	}
	if f.ranName != "osascript" {
		t.Errorf("Ask() ran %q", f.ranName)
	}
}

func TestAskRejectsEmptyPrompt(t *testing.T) {
	if _, err := Ask(&fakeExecer{}, Question{}); err == nil {
		t.Error("Ask() with empty prompt should error")
	}
}

func TestAskCancelled(t *testing.T) {
	f := &fakeExecer{outputErr: errors.New("execution error: User canceled. (-128)")}
	if _, err := Ask(f, Question{Prompt: "q"}); err == nil {
		t.Error("Ask() should fail when the user cancels")
	}
}
