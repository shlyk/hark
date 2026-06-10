// Package notify builds and runs macOS notification and speech commands.
// User-supplied text is always passed to osascript as argv, never
// interpolated into AppleScript source.
package notify

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Execer runs external commands; tests inject a fake.
type Execer interface {
	Run(name string, args ...string) error
	LookPath(name string) (string, error)
}

// SystemExecer runs real commands, forwarding their stderr.
type SystemExecer struct{}

func (SystemExecer) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (SystemExecer) LookPath(name string) (string, error) { return exec.LookPath(name) }

type Notification struct {
	Message  string
	Title    string
	Subtitle string
	Sound    string
}

// Script returns AppleScript that reads every value from argv. Clauses for
// subtitle/sound are included only when set, but argv positions are fixed.
func Script(n Notification) string {
	clause := "display notification (item 1 of argv) with title (item 2 of argv)"
	if n.Subtitle != "" {
		clause += " subtitle (item 3 of argv)"
	}
	if n.Sound != "" {
		clause += " sound name (item 4 of argv)"
	}
	return "on run argv\n" + clause + "\nend run"
}

// Args returns the full osascript argument list for n.
func Args(n Notification) []string {
	return []string{"-e", Script(n), "--", n.Message, n.Title, n.Subtitle, n.Sound}
}

// Send delivers a notification banner via osascript.
func Send(e Execer, n Notification) error {
	if strings.TrimSpace(n.Message) == "" {
		return fmt.Errorf("message must not be empty")
	}
	if _, err := e.LookPath("osascript"); err != nil {
		return fmt.Errorf("osascript not found (is this macOS?): %w", err)
	}
	if err := e.Run("osascript", Args(n)...); err != nil {
		return fmt.Errorf("notification delivery failed: %w", err)
	}
	return nil
}

type Speech struct {
	Text  string
	Voice string
	Rate  int
}

// SayArgs returns the argument list for the macOS say command.
func SayArgs(s Speech) []string {
	var args []string
	if s.Voice != "" {
		args = append(args, "-v", s.Voice)
	}
	if s.Rate > 0 {
		args = append(args, "-r", fmt.Sprint(s.Rate))
	}
	return append(args, "--", s.Text)
}

// Say speaks text aloud via the macOS say command.
func Say(e Execer, s Speech) error {
	if strings.TrimSpace(s.Text) == "" {
		return fmt.Errorf("text must not be empty")
	}
	if _, err := e.LookPath("say"); err != nil {
		return fmt.Errorf("say not found (is this macOS?): %w", err)
	}
	if err := e.Run("say", SayArgs(s)...); err != nil {
		return fmt.Errorf("speech failed: %w", err)
	}
	return nil
}
