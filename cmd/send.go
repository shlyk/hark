package cmd

import (
	"strings"

	"hark/internal/notify"

	"github.com/spf13/cobra"
)

func newSendCmd() *cobra.Command {
	var title, subtitle, sound string
	var speak bool
	cmd := &cobra.Command{
		Use:   "send <message>",
		Short: "Send a macOS notification banner",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			msg := strings.Join(args, " ")
			n := notify.Notification{Message: msg, Title: title, Subtitle: subtitle, Sound: sound}
			if err := notify.Send(execer, n); err != nil {
				return err
			}
			if speak {
				if err := notify.Say(execer, notify.Speech{Text: msg}); err != nil {
					return err
				}
			}
			record("send", title, msg)
			return nil
		},
	}
	cmd.Flags().StringVarP(&title, "title", "t", "hark", "notification title")
	cmd.Flags().StringVarP(&subtitle, "subtitle", "s", "", "notification subtitle")
	cmd.Flags().StringVar(&sound, "sound", "", `sound name, e.g. "Glass"`)
	cmd.Flags().BoolVar(&speak, "say", false, "also speak the message aloud")
	return cmd
}
