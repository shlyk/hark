package cmd

import (
	"fmt"

	"github.com/shlyk/hark/internal/config"
	"github.com/shlyk/hark/internal/notify"
	"github.com/shlyk/hark/internal/presence"
	"github.com/shlyk/hark/internal/remote"

	"github.com/spf13/cobra"
)

// delivery is the post-flag notification pipeline shared by send and relay:
// banner, history record, optional speech, optional or escalated remote push.
type delivery struct {
	kind         string // history kind: "send" or "relay"
	notification notify.Notification
	speak        bool   // force speech
	smart        bool   // speech only with headphones
	push         bool   // force remote push (failure = error)
	once         string // dedupe key already checked by the caller
}

func deliver(cmd *cobra.Command, execer notify.Execer, cfg config.Config, d delivery) error {
	if err := notify.Send(execer, d.notification); err != nil {
		return err
	}
	// The banner is delivered at this point — record it even if the
	// optional steps below fail.
	record(cmd, d.kind, d.notification.Title, d.notification.Message, d.once)
	if d.speak || (d.smart && notify.HeadphonesConnected(execer)) {
		if err := notify.Say(execer, notify.Speech{Text: d.notification.Message}); err != nil {
			return err
		}
	}
	ntfy := remote.Client{Server: cfg.Ntfy.ServerOrDefault(), Topic: cfg.Ntfy.Topic}
	if d.push {
		return ntfy.Send(d.notification.Title, d.notification.Message)
	}
	if cfg.Escalate.Enabled && cfg.Ntfy.Topic != "" &&
		presence.Away(execer, cfg.Escalate.IdleOrDefault()) {
		// Best-effort: the banner is already out, so escalation failure
		// only warns.
		if err := ntfy.Send(d.notification.Title, d.notification.Message); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", err)
		}
	}
	return nil
}
