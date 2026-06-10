package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/shlyk/hark/internal/config"
	"github.com/shlyk/hark/internal/history"
	"github.com/shlyk/hark/internal/notify"
	"github.com/shlyk/hark/internal/presence"

	"github.com/spf13/cobra"
)

func newDoctorCmd(execer notify.Execer) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check that hark can deliver notifications",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			failed := false
			check := func(name string, err error) {
				if err != nil {
					failed = true
					fmt.Fprintf(out, "FAIL %s: %v\n", name, err)
				} else {
					fmt.Fprintf(out, "ok   %s\n", name)
				}
			}

			_, err := execer.LookPath("osascript")
			check("osascript available", err)
			_, err = execer.LookPath("say")
			check("say available", err)

			store, err := history.DefaultStore()
			if err == nil {
				err = store.Append(history.Entry{Time: time.Now(), Kind: "doctor", Message: "doctor check"})
			}
			check("history writable", err)

			cfg, err := config.Load()
			check("config readable", err)
			if cfg.Ntfy.Topic != "" {
				fmt.Fprintf(out, "ok   ntfy configured (%s)\n", cfg.Ntfy.ServerOrDefault())
			} else {
				fmt.Fprintln(out, "note ntfy not configured — remote push and escalation are off")
			}

			idle, idleErr := presence.IdleSeconds(execer)
			name := "presence detection"
			if idleErr == nil {
				name = fmt.Sprintf("presence detection (idle %.0fs)", idle)
			}
			check(name, idleErr)

			if focusMayBeActive() {
				fmt.Fprintln(out, "note a Focus mode may be active — banners can be suppressed")
			}

			err = notify.Send(execer, notify.Notification{Message: "hark is working", Title: "hark doctor"})
			check("test notification sent", err)

			if failed {
				return fmt.Errorf("some checks failed")
			}
			fmt.Fprintln(out, "All good. If no banner appeared, allow Script Editor in System Settings > Notifications.")
			return nil
		},
	}
}

// focusMayBeActive is a loose heuristic: macOS records active Focus
// assertions in ~/Library/DoNotDisturb/DB/Assertions.json.
func focusMayBeActive() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	data, err := os.ReadFile(filepath.Join(home, "Library", "DoNotDisturb", "DB", "Assertions.json"))
	return err == nil && len(bytes.TrimSpace(data)) > 2 && bytes.Contains(data, []byte("assert"))
}
