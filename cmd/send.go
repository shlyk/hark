package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/shlyk/hark/internal/config"
	"github.com/shlyk/hark/internal/history"
	"github.com/shlyk/hark/internal/notify"
	"github.com/shlyk/hark/internal/presence"
	"github.com/shlyk/hark/internal/remote"

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
			smart = smart || cfg.Smart
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
			n := notify.Notification{Message: msg, Title: title, Subtitle: subtitle, Sound: sound}
			if err := notify.Send(execer, n); err != nil {
				return err
			}
			// The banner is delivered at this point — record it even if the
			// optional steps below fail.
			record(cmd, "send", title, msg, once)
			if speak || (smart && notify.HeadphonesConnected(execer)) {
				if err := notify.Say(execer, notify.Speech{Text: msg}); err != nil {
					return err
				}
			}
			ntfy := remote.Client{Server: cfg.Ntfy.ServerOrDefault(), Topic: cfg.Ntfy.Topic}
			if push {
				return ntfy.Send(title, msg)
			}
			if cfg.Escalate.Enabled && cfg.Ntfy.Topic != "" &&
				presence.Away(execer, cfg.Escalate.IdleOrDefault()) {
				// Best-effort: the banner is already out, so escalation
				// failure only warns.
				if err := ntfy.Send(title, msg); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", err)
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
	cmd.Flags().BoolVar(&push, "remote", false, "also push to the configured ntfy topic")
	cmd.Flags().StringVar(&once, "once", "", "dedupe key: skip if already sent with this key in the last 10 minutes")
	return cmd
}
