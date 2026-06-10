package presence

import (
	"errors"
	"testing"
)

type fakeExecer struct {
	output    []byte
	outputErr error
}

func (f *fakeExecer) Run(name string, args ...string) error { return nil }
func (f *fakeExecer) Output(name string, args ...string) ([]byte, error) {
	return f.output, f.outputErr
}
func (f *fakeExecer) LookPath(name string) (string, error) { return "/usr/sbin/" + name, nil }

func TestIdleSeconds(t *testing.T) {
	f := &fakeExecer{output: []byte(`    | |     "HIDIdleTime" = 8924352708`)}
	got, err := IdleSeconds(f)
	if err != nil {
		t.Fatalf("IdleSeconds() error = %v", err)
	}
	if got < 8.9 || got > 9.0 {
		t.Errorf("IdleSeconds() = %v, want ~8.92", got)
	}
}

func TestIdleSecondsNoMatch(t *testing.T) {
	f := &fakeExecer{output: []byte("nothing here")}
	if _, err := IdleSeconds(f); err == nil {
		t.Error("IdleSeconds() without HIDIdleTime should error")
	}
}

func TestScreenLocked(t *testing.T) {
	cases := []struct {
		output []byte
		err    error
		want   bool
	}{
		{[]byte(`  "IOConsoleLocked" = Yes`), nil, true},
		{[]byte(`  "IOConsoleLocked" = No`), nil, false},
		{[]byte("no such key"), nil, false},
		{nil, errors.New("boom"), false},
	}
	for _, c := range cases {
		f := &fakeExecer{output: c.output, outputErr: c.err}
		if got := ScreenLocked(f); got != c.want {
			t.Errorf("ScreenLocked(%q) = %v, want %v", c.output, got, c.want)
		}
	}
}

func TestAway(t *testing.T) {
	locked := &fakeExecer{output: []byte(`"IOConsoleLocked" = Yes`)}
	if !Away(locked, 300) {
		t.Error("Away() with locked screen should be true")
	}
	// Unlocked and HIDIdleTime below threshold (10s idle < 300s threshold).
	active := &fakeExecer{output: []byte(`"IOConsoleLocked" = No
"HIDIdleTime" = 10000000000`)}
	if Away(active, 300) {
		t.Error("Away() while active should be false")
	}
	// Unlocked but idle 400s > 300s threshold.
	idle := &fakeExecer{output: []byte(`"IOConsoleLocked" = No
"HIDIdleTime" = 400000000000`)}
	if !Away(idle, 300) {
		t.Error("Away() when idle beyond threshold should be true")
	}
	// Detection failure degrades to present.
	broken := &fakeExecer{outputErr: errors.New("boom")}
	if Away(broken, 300) {
		t.Error("Away() on detection failure should be false (assume present)")
	}
}
