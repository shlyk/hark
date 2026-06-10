package cmd

import (
	"fmt"
	"time"

	"hark/internal/history"
	"hark/internal/notify"

	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
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
