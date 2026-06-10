package cmd

import (
	"hark/internal/notify"

	"github.com/spf13/cobra"
)

// execer is the command runner used by all commands; tests replace it.
var execer notify.Execer = notify.SystemExecer{}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "hark",
		Short: "Get the user's attention with macOS notifications",
		Long: `hark sends native macOS notifications and optionally speaks them aloud.
It is designed for AI agents that need to notify a human, e.g.:

  hark send "Build finished" --title "CI" --sound Glass
  hark send "Need your input on the migration plan" --say

Every notification is recorded; see "hark history".`,
		SilenceUsage: true,
	}
	root.AddCommand(newSendCmd(), newSayCmd(), newDoctorCmd(), newHistoryCmd())
	return root
}

func Execute() error { return newRootCmd().Execute() }
