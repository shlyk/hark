package cmd

import (
	"strings"

	"github.com/shlyk/hark/internal/notify"

	"github.com/spf13/cobra"
)

func newSayCmd() *cobra.Command {
	var voice string
	var rate int
	cmd := &cobra.Command{
		Use:   "say <text>",
		Short: "Speak text aloud (no banner)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := strings.Join(args, " ")
			if err := notify.Say(execer, notify.Speech{Text: text, Voice: voice, Rate: rate}); err != nil {
				return err
			}
			record("say", "", text)
			return nil
		},
	}
	cmd.Flags().StringVar(&voice, "voice", "", "voice name, e.g. Samantha")
	cmd.Flags().IntVar(&rate, "rate", 0, "speech rate in words per minute")
	return cmd
}
