package cmd

import (
	"strings"

	"github.com/shlyk/hark/internal/config"
	"github.com/shlyk/hark/internal/notify"

	"github.com/spf13/cobra"
)

func newSendCmd(execer notify.Execer) *cobra.Command {
	var title, subtitle, sound string
	var speak, smart bool
	cmd := &cobra.Command{
		Use:   "send <message>",
		Short: "Send a macOS notification banner",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if !cmd.Flags().Changed("title") && cfg.Title != "" {
				title = cfg.Title
			}
			if !cmd.Flags().Changed("sound") && cfg.Sound != "" {
				sound = cfg.Sound
			}
			smart = smart || cfg.Smart
			msg := strings.Join(args, " ")
			n := notify.Notification{Message: msg, Title: title, Subtitle: subtitle, Sound: sound}
			if err := notify.Send(execer, n); err != nil {
				return err
			}
			// The banner is delivered at this point — record it even if the
			// optional speech below fails.
			record(cmd, "send", title, msg)
			if speak || (smart && notify.HeadphonesConnected(execer)) {
				if err := notify.Say(execer, notify.Speech{Text: msg}); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&title, "title", "t", "hark", "notification title")
	cmd.Flags().StringVarP(&subtitle, "subtitle", "s", "", "notification subtitle")
	cmd.Flags().StringVar(&sound, "sound", "", `sound name, e.g. "Glass"`)
	cmd.Flags().BoolVar(&speak, "say", false, "also speak the message aloud")
	cmd.Flags().BoolVar(&smart, "smart", false, "speak the message only when headphones are connected")
	return cmd
}
