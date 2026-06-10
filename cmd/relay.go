package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/shlyk/hark/internal/notify"

	"github.com/spf13/cobra"
)

// hookPayload is the subset of an agent hook event hark cares about.
// Claude Code sends {"hook_event_name": "...", "message": "...", ...}.
type hookPayload struct {
	HookEventName string `json:"hook_event_name"`
	Message       string `json:"message"`
	Title         string `json:"title"`
}

func newRelayCmd(execer notify.Execer) *cobra.Command {
	return &cobra.Command{
		Use:   "relay <agent>",
		Short: "Relay an agent hook payload from stdin into a notification",
		Long: `relay reads a hook event JSON payload from stdin and turns it into a
smart notification (spoken with headphones, escalated to ntfy when away).
It is meant to be wired as a hook command — see "hark hook claude".`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var p hookPayload
			// Empty stdin degrades to the default message: relay's job is
			// to ping the user, so only malformed JSON is an error.
			if err := json.NewDecoder(cmd.InOrStdin()).Decode(&p); err != nil && err != io.EOF {
				return fmt.Errorf("reading hook payload from stdin: %w", err)
			}
			agent := args[0]
			msg := p.Message
			if msg == "" {
				switch p.HookEventName {
				case "Stop":
					msg = agent + " finished and is waiting"
				default:
					msg = agent + " needs your attention"
				}
			}
			title := p.Title
			if title == "" {
				title = agent
			}
			return deliver(cmd, execer, delivery{
				kind:         "relay",
				notification: notify.Notification{Message: msg, Title: title},
				smart:        true,
			})
		},
	}
}
