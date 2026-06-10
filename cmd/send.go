package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/shlyk/hark/internal/config"
	"github.com/shlyk/hark/internal/history"
	"github.com/shlyk/hark/internal/notify"

	"github.com/spf13/cobra"
)

// onceWindow is how long a --once dedupe key suppresses repeats.
const onceWindow = 10 * time.Minute

func newSendCmd(execer notify.Execer) *cobra.Command {
	var title, subtitle, sound, once string
	var speak, smart, push bool
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
			msg := strings.Join(args, " ")
			if once != "" {
				store, err := history.DefaultStore()
				if err == nil {
					dup, err := store.HasRecent(once, time.Now().Add(-onceWindow))
					if err == nil && dup {
						fmt.Fprintf(cmd.OutOrStdout(), "skipped: %q already sent within %s\n", once, onceWindow)
						return nil
					}
				}
			}
			return deliver(cmd, execer, cfg, delivery{
				kind:         "send",
				notification: notify.Notification{Message: msg, Title: title, Subtitle: subtitle, Sound: sound},
				speak:        speak,
				smart:        smart || cfg.Smart,
				push:         push,
				once:         once,
			})
		},
	}
	cmd.Flags().StringVarP(&title, "title", "t", "hark", "notification title")
	cmd.Flags().StringVarP(&subtitle, "subtitle", "s", "", "notification subtitle")
	cmd.Flags().StringVar(&sound, "sound", "", `sound name, e.g. "Glass"`)
	cmd.Flags().BoolVar(&speak, "say", false, "also speak the message aloud")
	cmd.Flags().BoolVar(&smart, "smart", false, "speak the message only when headphones are connected")
	cmd.Flags().BoolVar(&push, "remote", false, "also push to the configured ntfy topic")
	cmd.Flags().StringVar(&once, "once", "", "dedupe key: skip if already sent with this key in the last 10 minutes")
	return cmd
}
