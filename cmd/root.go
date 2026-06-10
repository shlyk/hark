package cmd

import (
	"github.com/shlyk/hark/internal/notify"

	"github.com/spf13/cobra"
)

// version is stamped by the release build:
// go build -ldflags "-X github.com/shlyk/hark/cmd.version=v1.2.3"
var version = "dev"

func newRootCmd(execer notify.Execer) *cobra.Command {
	root := &cobra.Command{
		Use:   "hark",
		Short: "Get the user's attention with macOS notifications",
		Long: `hark sends native macOS notifications and optionally speaks them aloud.
It is designed for AI agents that need to notify a human, e.g.:

  hark send "Build finished" --title "CI" --sound Glass
  hark send "Need your input on the migration plan" --say

Every notification is recorded; see "hark history".`,
		Version:      version,
		SilenceUsage: true,
	}
	root.AddCommand(newSendCmd(execer), newSayCmd(execer), newAskCmd(execer), newDoctorCmd(execer),
		newHistoryCmd(), newSkillCmd(), newRelayCmd(execer), newHookCmd())
	return root
}

func Execute() error { return newRootCmd(notify.SystemExecer{}).Execute() }
