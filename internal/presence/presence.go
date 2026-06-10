// Package presence estimates whether the user is at the Mac, via ioreg.
// All checks are best-effort: on any failure callers should assume the
// user is present (no escalation) rather than erroring.
package presence

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/shlyk/hark/internal/notify"
)

var (
	idleRe   = regexp.MustCompile(`"HIDIdleTime"\s*=\s*(\d+)`)
	lockedRe = regexp.MustCompile(`"IOConsoleLocked"\s*=\s*Yes`)
)

// IdleSeconds returns seconds since the last keyboard/mouse input.
func IdleSeconds(e notify.Execer) (float64, error) {
	out, err := e.Output("ioreg", "-c", "IOHIDSystem", "-d", "4")
	if err != nil {
		return 0, err
	}
	m := idleRe.FindSubmatch(out)
	if m == nil {
		return 0, fmt.Errorf("HIDIdleTime not found in ioreg output")
	}
	ns, err := strconv.ParseFloat(string(m[1]), 64)
	if err != nil {
		return 0, err
	}
	return ns / 1e9, nil
}

// ScreenLocked reports whether the console is locked; false on any failure.
func ScreenLocked(e notify.Execer) bool {
	out, err := e.Output("ioreg", "-n", "Root", "-d", "1")
	if err != nil {
		return false
	}
	return lockedRe.Match(out)
}

// Away reports whether the user looks away from the Mac: screen locked, or
// no input for more than idleThreshold seconds.
func Away(e notify.Execer, idleThreshold int) bool {
	if ScreenLocked(e) {
		return true
	}
	idle, err := IdleSeconds(e)
	if err != nil {
		return false
	}
	return idle > float64(idleThreshold)
}
