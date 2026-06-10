package notify

import (
	"fmt"
	"strings"
)

// Question describes an interactive dialog. Zero Options and Input=false
// yields a Yes/No dialog. Timeout (seconds) applies to dialogs only:
// AppleScript's "choose from list" cannot give up.
type Question struct {
	Prompt  string
	Title   string
	Options []string
	Input   bool
	Timeout int
}

// AskScript returns AppleScript reading all user text from argv (same
// injection-safety property as Script). The numeric timeout is embedded
// directly; it never contains user text.
func AskScript(q Question) string {
	givingUp := ""
	if q.Timeout > 0 {
		givingUp = fmt.Sprintf(" giving up after %d", q.Timeout)
	}
	var body string
	switch {
	case len(q.Options) > 0:
		body = `set choice to choose from list (items 3 thru -1 of argv) with prompt (item 1 of argv) with title (item 2 of argv)
if choice is false then error number -128
return item 1 of choice`
	case q.Input:
		body = `set r to display dialog (item 1 of argv) with title (item 2 of argv) default answer ""` + givingUp + `
if gave up of r then error number -1712
return text returned of r`
	default:
		body = `set r to display dialog (item 1 of argv) with title (item 2 of argv) buttons {"No", "Yes"} default button "Yes"` + givingUp + `
if gave up of r then error number -1712
return button returned of r`
	}
	return "on run argv\n" + body + "\nend run"
}

// AskArgs returns the full osascript argument list for q.
func AskArgs(q Question) []string {
	args := []string{"-e", AskScript(q), "--", q.Prompt, q.Title}
	return append(args, q.Options...)
}

// Ask shows the dialog and returns the user's answer (chosen option,
// pressed button, or typed text). Cancel and timeout are errors.
func Ask(e Execer, q Question) (string, error) {
	if strings.TrimSpace(q.Prompt) == "" {
		return "", fmt.Errorf("question must not be empty")
	}
	if _, err := e.LookPath("osascript"); err != nil {
		return "", fmt.Errorf("osascript not found (is this macOS?): %w", err)
	}
	out, err := e.Output("osascript", AskArgs(q)...)
	if err != nil {
		return "", fmt.Errorf("no answer (cancelled, timed out, or dialog failed): %w", err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}
